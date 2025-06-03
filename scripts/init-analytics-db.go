package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./db/chat_widget.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create chat_metrics table
	createChatMetricsTable := `
	CREATE TABLE IF NOT EXISTS chat_metrics (
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
	);`

	_, err = db.Exec(createChatMetricsTable)
	if err != nil {
		log.Fatal("Failed to create chat_metrics table:", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_chat_metrics_customer_id ON chat_metrics(customer_id);",
		"CREATE INDEX IF NOT EXISTS idx_chat_metrics_session_id ON chat_metrics(session_id);",
		"CREATE INDEX IF NOT EXISTS idx_chat_metrics_created_at ON chat_metrics(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_chat_metrics_country ON chat_metrics(country);",
		"CREATE INDEX IF NOT EXISTS idx_chat_metrics_device_type ON chat_metrics(device_type);",
	}

	for _, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	log.Println("Analytics database tables created successfully!")
}
