package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
)

// CustomerService handles customer-related operations
type CustomerService struct {
	db *sql.DB
}

// NewCustomerService creates a new customer service
func NewCustomerService(db *sql.DB) *CustomerService {
	return &CustomerService{db: db}
}

// GetByID retrieves a customer by ID
func (s *CustomerService) GetByID(id string) (*models.Customer, error) {
	var customer models.Customer
	err := s.db.QueryRow(`
		SELECT id, name, email, created_at, brand_colors, logo_url, model_path,
		       animations, openai_prompt, allowed_domains, api_key, active
		FROM customers WHERE id = ?
	`, id).Scan(
		&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt,
		&customer.BrandColors, &customer.LogoURL, &customer.ModelPath,
		&customer.Animations, &customer.OpenAIPrompt, &customer.AllowedDomains,
		&customer.APIKey, &customer.Active,
	)

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// GetActiveByID retrieves an active customer by ID
func (s *CustomerService) GetActiveByID(id string) (*models.Customer, error) {
	var customer models.Customer
	err := s.db.QueryRow(`
		SELECT id, name, email, created_at, brand_colors, logo_url, model_path,
		       animations, openai_prompt, allowed_domains, api_key, active
		FROM customers WHERE id = ? AND active = 1
	`, id).Scan(
		&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt,
		&customer.BrandColors, &customer.LogoURL, &customer.ModelPath,
		&customer.Animations, &customer.OpenAIPrompt, &customer.AllowedDomains,
		&customer.APIKey, &customer.Active,
	)

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// GetAll retrieves all customers
func (s *CustomerService) GetAll() ([]models.Customer, error) {
	rows, err := s.db.Query(`
		SELECT id, name, email, created_at, brand_colors, logo_url, model_path,
		       animations, openai_prompt, allowed_domains, api_key, active
		FROM customers ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(
			&c.ID, &c.Name, &c.Email, &c.CreatedAt,
			&c.BrandColors, &c.LogoURL, &c.ModelPath,
			&c.Animations, &c.OpenAIPrompt, &c.AllowedDomains,
			&c.APIKey, &c.Active,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	return customers, nil
}

// GetTotalCount returns the total number of customers
func (s *CustomerService) GetTotalCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM customers WHERE active = 1").Scan(&count)
	return count, err
}

// Create creates a new customer
func (s *CustomerService) Create(customer *models.Customer) error {
	// Generate ID and API key
	customer.ID = generateID()
	customer.APIKey = generateAPIKey()
	customer.CreatedAt = time.Now()

	// Set default brand colors if empty
	if customer.BrandColors == "" {
		customer.BrandColors = "primary: #007bff\nsecondary: #6c757d\nbackground: #ffffff\ntext: #212529"
	}

	_, err := s.db.Exec(`
		INSERT INTO customers (
			id, name, email, created_at, brand_colors, logo_url, model_path,
			animations, openai_prompt, allowed_domains, api_key, active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		customer.ID, customer.Name, customer.Email, customer.CreatedAt,
		customer.BrandColors, customer.LogoURL, customer.ModelPath,
		customer.Animations, customer.OpenAIPrompt, customer.AllowedDomains,
		customer.APIKey, customer.Active,
	)
	return err
}

// Update updates a customer
func (s *CustomerService) Update(id string, customer *models.Customer) error {
	_, err := s.db.Exec(`
		UPDATE customers SET
			name = ?, email = ?, brand_colors = ?, logo_url = ?,
			openai_prompt = ?, allowed_domains = ?, active = ?
		WHERE id = ?
	`,
		customer.Name, customer.Email, customer.BrandColors, customer.LogoURL,
		customer.OpenAIPrompt, customer.AllowedDomains, customer.Active, id,
	)
	return err
}

// UpdateModelPath updates the model path for a customer
func (s *CustomerService) UpdateModelPath(id, modelPath string) error {
	_, err := s.db.Exec("UPDATE customers SET model_path = ? WHERE id = ?", modelPath, id)
	return err
}

// ValidateOrigin checks if the origin is allowed for the customer
func (s *CustomerService) ValidateOrigin(origin, customerID string) bool {
	var allowedDomains string
	err := s.db.QueryRow("SELECT allowed_domains FROM customers WHERE id = ?", customerID).
		Scan(&allowedDomains)

	if err != nil {
		return false
	}

	// If no domains are configured, reject the origin
	if strings.TrimSpace(allowedDomains) == "" {
		return false
	}

	// Check if origin matches any allowed domain
	domains := strings.Split(allowedDomains, ",")
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		if domain == "*" || strings.Contains(origin, domain) {
			return true
		}
	}

	return false
}

// ChatService handles chat-related operations
type ChatService struct {
	db *sql.DB
}

// NewChatService creates a new chat service
func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{db: db}
}

// CreateSession creates a new chat session
func (s *ChatService) CreateSession(customerID, userID string) (string, error) {
	sessionID := generateID()
	_, err := s.db.Exec(`
		INSERT INTO chat_sessions (id, customer_id, user_id, metadata)
		VALUES (?, ?, ?, ?)
	`, sessionID, customerID, userID, "{}")

	return sessionID, err
}

// EndSession ends a chat session
func (s *ChatService) EndSession(sessionID string) error {
	_, err := s.db.Exec("UPDATE chat_sessions SET ended_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
	return err
}

// CreateMessage creates a new chat message
func (s *ChatService) CreateMessage(sessionID, role, content, emotion string) (string, error) {
	messageID := generateID()
	_, err := s.db.Exec(`
		INSERT INTO chat_messages (id, session_id, role, content, emotion)
		VALUES (?, ?, ?, ?, ?)
	`, messageID, sessionID, role, content, emotion)

	return messageID, err
}

// GetSessions retrieves chat sessions with optional customer filter
func (s *ChatService) GetSessions(customerID string, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT s.id, s.user_id, s.started_at, s.ended_at, 
		       COUNT(m.id) as message_count
		FROM chat_sessions s
		LEFT JOIN chat_messages m ON s.id = m.session_id
		WHERE 1=1
	`
	args := []interface{}{}

	if customerID != "" {
		query += " AND s.customer_id = ?"
		args = append(args, customerID)
	}

	query += " GROUP BY s.id ORDER BY s.started_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var sessions []map[string]interface{}
	for rows.Next() {
		var session models.ChatSession
		var messageCount int
		err := rows.Scan(&session.ID, &session.UserID, &session.StartedAt, &session.EndedAt, &messageCount)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, map[string]interface{}{
			"ID":           session.ID,
			"UserID":       session.UserID,
			"StartedAt":    session.StartedAt,
			"EndedAt":      session.EndedAt,
			"MessageCount": messageCount,
		})
	}

	return sessions, nil
}

// OpenAIService handles OpenAI integration
type OpenAIService struct{}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService() *OpenAIService {
	return &OpenAIService{}
}

// GenerateResponse generates a response using OpenAI (mock implementation)
func (s *OpenAIService) GenerateResponse(prompt, message string) (string, string) {
	// Mock implementation - replace with actual OpenAI API call
	emotions := []string{"happy", "thinking", "neutral", "excited"}
	emotion := emotions[time.Now().Unix()%int64(len(emotions))]

	response := fmt.Sprintf("Response to: %s (with prompt: %s)", message, prompt[:min(20, len(prompt))])
	return response, emotion
}

// AuthService handles authentication
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// GetAdminUser retrieves an admin user by username
func (s *AuthService) GetAdminUser(username string) (*models.AdminUser, error) {
	var user models.AdminUser
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM admin_users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateAPIKey() string {
	return fmt.Sprintf("api_%d", time.Now().UnixNano())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
