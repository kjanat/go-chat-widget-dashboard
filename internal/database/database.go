package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps sql.DB to provide additional methods
type DB struct {
	*sql.DB
}

// New creates a new database connection and initializes tables
func New(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbWrapper := &DB{DB: db}
	if err := dbWrapper.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbWrapper, nil
}

// createTables creates all necessary tables
func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			brand_colors TEXT,
			logo_url TEXT,
			model_path TEXT,
			animations TEXT,
			openai_prompt TEXT,
			allowed_domains TEXT,
			api_key TEXT UNIQUE,
			active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ended_at DATETIME,
			metadata TEXT,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			emotion TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES chat_sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_metrics (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			country TEXT,
			city TEXT,
			session_start_time DATETIME NOT NULL,
			session_end_time DATETIME,
			message_count INTEGER DEFAULT 0,
			response_time INTEGER DEFAULT 0,
			satisfaction_score INTEGER,
			conversion_event TEXT,
			page_url TEXT,
			referrer_url TEXT,
			device_type TEXT,
			browser_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers (id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_customer ON chat_sessions(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON chat_messages(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_metrics_customer_id ON chat_metrics(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_metrics_session_id ON chat_metrics(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_metrics_created_at ON chat_metrics(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_metrics_country ON chat_metrics(country)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_metrics_device_type ON chat_metrics(device_type)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}
