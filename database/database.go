package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	dbPath := "./database/debt.note.db"

	// Ensure the directory exists
	if _, err := os.Stat("./database"); os.IsNotExist(err) {
		os.Mkdir("./database", os.ModePerm)
	}

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable foreign key support
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	log.Println("Database connection successful.")
	createTables()
	migrateTables()
}

func createTables() {
	createClientsTableSQL := `CREATE TABLE IF NOT EXISTS clients (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"fullname" TEXT NOT NULL,
		"phone" TEXT NOT NULL UNIQUE,
		"address" TEXT,
		"photo_data" TEXT,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(createClientsTableSQL); err != nil {
		log.Fatalf("Failed to create clients table: %v", err)
	}
	log.Println("Clients table created or already exists.")

	createDebtsTableSQL := `CREATE TABLE IF NOT EXISTS debts (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"client_id" INTEGER NOT NULL,
		"amount" REAL NOT NULL,
		"comment" TEXT,
		"status" TEXT NOT NULL DEFAULT 'active',
		"rating" TEXT,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
		"paid_at" DATETIME,
		"deleted_at" DATETIME,
		"delete_comment" TEXT,
		FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(createDebtsTableSQL); err != nil {
		log.Fatalf("Failed to create debts table: %v", err)
	}
	log.Println("Debts table created or already exists.")

	createDebtPaymentsTableSQL := `CREATE TABLE IF NOT EXISTS debt_payments (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"debt_id" INTEGER NOT NULL,
		"paid_amount" REAL NOT NULL,
		"remaining_amount" REAL NOT NULL,
		"comment" TEXT,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(createDebtPaymentsTableSQL); err != nil {
		log.Fatalf("Failed to create debt_payments table: %v", err)
	}
	log.Println("Debt payments table created or already exists.")
}

func migrateTables() {
	// Try to add new columns if they don't exist.
	// SQLite doesn't support "ADD COLUMN IF NOT EXISTS", so we just try and ignore error.

	_, _ = DB.Exec(`ALTER TABLE debts ADD COLUMN deleted_at DATETIME;`)
	_, _ = DB.Exec(`ALTER TABLE debts ADD COLUMN delete_comment TEXT;`)
}
