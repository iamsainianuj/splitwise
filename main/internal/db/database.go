package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	var err error

	// Use DATABASE_URL for Turso, or local SQLite file
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "./splitwise.db"
	}

	// For Turso URLs, use libsql driver
	driverName := "sqlite3"
	if len(dbURL) > 8 && dbURL[:8] == "libsql://" {
		// For Turso, we need to convert the URL format
		driverName = "sqlite3"
		dbURL = dbURL + "?_auth_token=" + os.Getenv("DATABASE_AUTH_TOKEN")
	}

	DB, err = sql.Open(driverName, dbURL)
	if err != nil {
		return err
	}

	// Create tables
	if err := createTables(); err != nil {
		return err
	}

	log.Println("✅ Database initialized")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			user_name TEXT NOT NULL,
			user_email TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS groups (
			group_id TEXT PRIMARY KEY,
			group_name TEXT NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS group_members (
			group_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (group_id, user_id),
			FOREIGN KEY (group_id) REFERENCES groups(group_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS expenses (
			expense_id TEXT PRIMARY KEY,
			expense_description TEXT NOT NULL,
			expense_amount REAL NOT NULL,
			group_id TEXT NOT NULL,
			paid_by_user_id TEXT NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (group_id) REFERENCES groups(group_id),
			FOREIGN KEY (paid_by_user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS splits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expense_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			amount REAL NOT NULL,
			FOREIGN KEY (expense_id) REFERENCES expenses(expense_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS balances (
			from_user_id TEXT NOT NULL,
			to_user_id TEXT NOT NULL,
			amount REAL NOT NULL DEFAULT 0,
			PRIMARY KEY (from_user_id, to_user_id),
			FOREIGN KEY (from_user_id) REFERENCES users(user_id),
			FOREIGN KEY (to_user_id) REFERENCES users(user_id)
		)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

