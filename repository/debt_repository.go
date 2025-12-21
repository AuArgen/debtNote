package repository

import (
	"debtNote/database"
	"debtNote/models"
	"time"
)

// CombinedDebtInfo is a struct for joining client and debt info
type CombinedDebtInfo struct {
	DebtID    int64      `json:"debt_id"`
	ClientID  int64      `json:"client_id"`
	Fullname  string     `json:"fullname"`
	Phone     string     `json:"phone"`
	Address   string     `json:"address"` // Added Address field
	PhotoData string     `json:"photo_data"`
	Amount    float64    `json:"amount"`
	Comment   string     `json:"comment"`
	Status    string     `json:"status"`
	Rating    *string    `json:"rating"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at"`
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
		whereClause += " AND date(d.created_at) = ?"
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
	switch sortBy {
	case "date_old":
		orderBy = "d.created_at ASC, d.id ASC"
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
			d.amount, d.comment, d.status, d.rating, d.created_at, d.paid_at
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
		if err := rows.Scan(
			&d.DebtID, &d.ClientID, &d.Fullname, &d.Phone, &d.Address, &d.PhotoData,
			&d.Amount, &d.Comment, &d.Status, &d.Rating, &d.CreatedAt, &d.PaidAt,
		); err != nil {
			return nil, 0, err
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

// PayDebt marks a debt as paid and gives it a rating.
func PayDebt(debtID int64, rating models.DebtRating) error {
	stmt, err := database.DB.Prepare("UPDATE debts SET status = ?, rating = ?, paid_at = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(models.StatusPaid, rating, time.Now(), debtID)
	return err
}
