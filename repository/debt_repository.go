package repository

import (
	"debtNote/database"
	"debtNote/models"
	"time"
)

// CombinedDebtInfo is a struct for joining client and debt info
type CombinedDebtInfo struct {
	DebtID        int64      `json:"debt_id"`
	ClientID      int64      `json:"client_id"`
	Fullname      string     `json:"fullname"`
	Phone         string     `json:"phone"`
	Address       string     `json:"address"`
	PhotoData     string     `json:"photo_data"`
	Amount        float64    `json:"amount"`
	Comment       string     `json:"comment"`
	Status        string     `json:"status"`
	Rating        *string    `json:"rating"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	DeleteComment string     `json:"delete_comment"`
}

// GetDebts retrieves a list of debts based on filters, sorting, and pagination.
func GetDebts(search, date, status string, clientID int64, sortBy string, page, limit int) ([]CombinedDebtInfo, int, error) {
	offset := (page - 1) * limit

	// Base query conditions
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		whereClause += " AND d.status = ?"
		args = append(args, status)
	}

	if clientID > 0 {
		whereClause += " AND d.client_id = ?"
		args = append(args, clientID)
	}

	if search != "" {
		whereClause += " AND (c.fullname LIKE ? OR c.phone LIKE ? OR c.address LIKE ? OR d.comment LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if date != "" {
		// If status is deleted, filter by deleted_at, otherwise created_at
		if status == "deleted" {
			whereClause += " AND date(d.deleted_at) = ?"
		} else {
			whereClause += " AND date(d.created_at) = ?"
		}
		args = append(args, date)
	}

	// 1. Get Total Count
	countQuery := `
		SELECT COUNT(*)
		FROM debts d
		JOIN clients c ON d.client_id = c.id` + whereClause

	var totalCount int
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Determine Sorting
	orderBy := "d.created_at DESC, d.id DESC" // Default: Newest first
	if status == "deleted" {
		orderBy = "d.deleted_at DESC, d.id DESC"
	}

	switch sortBy {
	case "date_old":
		if status == "deleted" {
			orderBy = "d.deleted_at ASC, d.id ASC"
		} else {
			orderBy = "d.created_at ASC, d.id ASC"
		}
	case "name":
		orderBy = "c.fullname ASC"
	case "amount_desc":
		orderBy = "d.amount DESC"
	case "amount_asc":
		orderBy = "d.amount ASC"
	}

	// 2. Get Data
	query := `
		SELECT
			d.id, d.client_id, c.fullname, c.phone, c.address, c.photo_data,
			d.amount, d.comment, d.status, d.rating, d.created_at, d.paid_at, d.deleted_at, d.delete_comment
		FROM debts d
		JOIN clients c ON d.client_id = c.id` + whereClause + ` ORDER BY ` + orderBy + ` LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var debts []CombinedDebtInfo
	for rows.Next() {
		var d CombinedDebtInfo
		// Handle NULLs for optional fields
		var deleteComment *string

		if err := rows.Scan(
			&d.DebtID, &d.ClientID, &d.Fullname, &d.Phone, &d.Address, &d.PhotoData,
			&d.Amount, &d.Comment, &d.Status, &d.Rating, &d.CreatedAt, &d.PaidAt, &d.DeletedAt, &deleteComment,
		); err != nil {
			return nil, 0, err
		}
		if deleteComment != nil {
			d.DeleteComment = *deleteComment
		}
		debts = append(debts, d)
	}

	if debts == nil {
		debts = []CombinedDebtInfo{}
	}

	return debts, totalCount, nil
}

// AddDebt adds a new debt record for a specific client.
func AddDebt(debt models.Debt) (int64, error) {
	stmt, err := database.DB.Prepare("INSERT INTO debts(client_id, amount, comment) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(debt.ClientID, debt.Amount, debt.Comment)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// MakePayment processes a partial or full payment.
func MakePayment(debtID int64, paidAmount float64, comment string, rating models.DebtRating) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get current debt amount
	var currentAmount float64
	err = tx.QueryRow("SELECT amount FROM debts WHERE id = ?", debtID).Scan(&currentAmount)
	if err != nil {
		return err
	}

	remainingAmount := currentAmount - paidAmount
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	// 2. Record the payment
	_, err = tx.Exec("INSERT INTO debt_payments(debt_id, paid_amount, remaining_amount, comment) VALUES(?, ?, ?, ?)",
		debtID, paidAmount, remainingAmount, comment)
	if err != nil {
		return err
	}

	// 3. Update debt amount and status
	if remainingAmount <= 0 {
		// Full payment - Close the debt
		_, err = tx.Exec("UPDATE debts SET amount = 0, status = ?, rating = ?, paid_at = ? WHERE id = ?",
			models.StatusPaid, rating, time.Now(), debtID)
	} else {
		// Partial payment - Update amount only
		_, err = tx.Exec("UPDATE debts SET amount = ? WHERE id = ?", remainingAmount, debtID)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetDebtPayments retrieves payment history for a specific debt.
func GetDebtPayments(debtID int64) ([]models.DebtPayment, error) {
	query := "SELECT id, debt_id, paid_amount, remaining_amount, comment, created_at FROM debt_payments WHERE debt_id = ? ORDER BY created_at DESC"
	rows, err := database.DB.Query(query, debtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.DebtPayment
	for rows.Next() {
		var p models.DebtPayment
		if err := rows.Scan(&p.ID, &p.DebtID, &p.PaidAmount, &p.RemainingAmount, &p.Comment, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// PayDebt marks a debt as paid and gives it a rating (Legacy function, kept for compatibility but MakePayment is preferred).
func PayDebt(debtID int64, rating models.DebtRating) error {
	stmt, err := database.DB.Prepare("UPDATE debts SET status = ?, rating = ?, paid_at = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(models.StatusPaid, rating, time.Now(), debtID)
	return err
}

// DeleteDebt marks a debt as deleted (soft delete) with a comment and timestamp.
func DeleteDebt(debtID int64, comment string) error {
	stmt, err := database.DB.Prepare("UPDATE debts SET status = ?, deleted_at = ?, delete_comment = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(models.StatusDeleted, time.Now(), comment, debtID)
	return err
}
