package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/kjanat/go-chat-widget-dashboard/internal/database"
)

func main() {
	// Initialize database
	db, err := database.New("./db/chat_widget.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create default admin user
	username := "admin"
	password := "admin123"

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Check if admin user already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Fatal("Failed to check for existing admin user:", err)
	}

	if count > 0 {
		fmt.Printf("Admin user '%s' already exists\n", username)
		return
	}

	// Create admin user
	_, err = db.Exec(`
		INSERT INTO admin_users (id, username, password_hash)
		VALUES (?, ?, ?)
	`, fmt.Sprintf("admin-%d", 1), username, string(hashedPassword))

	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Printf("✅ Created admin user:\n")
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Password: %s\n", password)
	fmt.Printf("   Login URL: http://localhost:8080/admin/login\n")
	fmt.Printf("\n⚠️  Please change the default password after first login!\n")
}
