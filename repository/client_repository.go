package repository

import (
	"database/sql"
	"debtNote/database"
	"debtNote/models"
	"errors"
)

// FindOrCreateClient finds a client by phone number or creates a new one.
// It requires a photo for new clients.
func FindOrCreateClient(client models.Client) (int64, error) {
	// Check if client exists
	var clientID int64
	err := database.DB.QueryRow("SELECT id FROM clients WHERE phone = ?", client.Phone).Scan(&clientID)
	
	if err == sql.ErrNoRows {
		// Client does not exist, create new. Photo is mandatory.
		if client.PhotoData == "" {
			return 0, errors.New("жаңы клиент үчүн сүрөт милдеттүү")
		}

		stmt, err := database.DB.Prepare("INSERT INTO clients(fullname, phone, address, photo_data) VALUES(?, ?, ?, ?)")
		if err != nil {
			return 0, err
		}
		defer stmt.Close()

		res, err := stmt.Exec(client.Fullname, client.Phone, client.Address, client.PhotoData)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	} else if err != nil {
		return 0, err
	}

	// Client exists, return ID
	return clientID, nil
}

// SearchClients searches for clients, checks active debts, and calculates reputation.
func SearchClients(query string) ([]models.ClientSearchInfo, error) {
	sqlQuery := `
		SELECT
			c.id,
			c.fullname,
			c.phone,
			c.address,
			c.photo_data,
			-- Check if there is any active debt
			EXISTS(SELECT 1 FROM debts d WHERE d.client_id = c.id AND d.status = 'active') as has_active_debt,
			-- Calculate reputation based on worst rating
			(
				SELECT 
					CASE 
						WHEN COUNT(CASE WHEN rating = 'untrusted' THEN 1 END) > 0 THEN 'untrusted'
						WHEN COUNT(CASE WHEN rating = 'bad' THEN 1 END) > 0 THEN 'bad'
						WHEN COUNT(CASE WHEN rating = 'good' THEN 1 END) > 0 THEN 'good'
						ELSE 'none'
					END
				FROM debts d 
				WHERE d.client_id = c.id AND d.status = 'paid'
			) as reputation
		FROM clients c
		WHERE c.fullname LIKE ? OR c.phone LIKE ?
		GROUP BY c.id
		LIMIT 5;
	`
	
	rows, err := database.DB.Query(sqlQuery, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.ClientSearchInfo
	for rows.Next() {
		var c models.ClientSearchInfo
		// Handle NULL reputation (if no paid debts)
		var reputation sql.NullString
		
		if err := rows.Scan(&c.ID, &c.Fullname, &c.Phone, &c.Address, &c.PhotoData, &c.HasActiveDebt, &reputation); err != nil {
			return nil, err
		}
		if reputation.Valid {
			c.Reputation = reputation.String
		} else {
			c.Reputation = "none"
		}
		clients = append(clients, c)
	}
	return clients, nil
}

// GetClients retrieves a paginated list of clients with filters, returning data and total count.
func GetClients(search, date string, page, limit int) ([]models.Client, int, error) {
	offset := (page - 1) * limit
	
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if search != "" {
		whereClause += " AND (fullname LIKE ? OR phone LIKE ? OR address LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if date != "" {
		whereClause += " AND date(created_at) = ?"
		args = append(args, date)
	}

	// 1. Get Total Count
	countQuery := "SELECT COUNT(*) FROM clients" + whereClause
	var totalCount int
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// 2. Get Data
	query := "SELECT id, fullname, phone, address, photo_data, created_at FROM clients" + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.Fullname, &c.Phone, &c.Address, &c.PhotoData, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		clients = append(clients, c)
	}

	if clients == nil {
		clients = []models.Client{}
	}

	return clients, totalCount, nil
}
