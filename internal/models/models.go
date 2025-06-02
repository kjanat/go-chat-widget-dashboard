package models

import (
	"encoding/json"
	"strings"
	"time"
)

// Customer represents a client using the chat widget
type Customer struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Email          string    `json:"email" db:"email"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	BrandColors    string    `json:"brand_colors" db:"brand_colors"` // Simple key-value format
	LogoURL        string    `json:"logo_url" db:"logo_url"`
	ModelPath      string    `json:"model_path" db:"model_path"`
	Animations     string    `json:"animations" db:"animations"` // JSON string
	OpenAIPrompt   string    `json:"openai_prompt" db:"openai_prompt"`
	AllowedDomains string    `json:"allowed_domains" db:"allowed_domains"` // Comma-separated
	APIKey         string    `json:"api_key" db:"api_key"`
	Active         bool      `json:"active" db:"active"`
}

// BrandColorsToJSON converts simple key-value format to JSON for the widget
func (c *Customer) BrandColorsToJSON() string {
	if c.BrandColors == "" {
		return `{"primary": "#007bff", "secondary": "#6c757d", "background": "#ffffff", "text": "#212529"}`
	}

	colors := make(map[string]string)
	
	// Handle both literal \n and actual newlines
	text := strings.ReplaceAll(c.BrandColors, "\\n", "\n")
	
	// Parse simple format: primary: #007bff\nsecondary: #6c757d
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			colors[key] = value
		}
	}
	
	// Set defaults if missing
	if _, exists := colors["primary"]; !exists {
		colors["primary"] = "#007bff"
	}
	if _, exists := colors["secondary"]; !exists {
		colors["secondary"] = "#6c757d"
	}
	if _, exists := colors["background"]; !exists {
		colors["background"] = "#ffffff"
	}
	if _, exists := colors["text"]; !exists {
		colors["text"] = "#212529"
	}
	
	jsonBytes, _ := json.Marshal(colors)
	return string(jsonBytes)
}

// SetBrandColorsFromJSON converts JSON to simple format (for backwards compatibility)
func (c *Customer) SetBrandColorsFromJSON(jsonStr string) {
	if jsonStr == "" {
		c.BrandColors = "primary: #007bff\nsecondary: #6c757d\nbackground: #ffffff\ntext: #212529"
		return
	}
	
	var colors map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &colors); err != nil {
		c.BrandColors = "primary: #007bff\nsecondary: #6c757d\nbackground: #ffffff\ntext: #212529"
		return
	}
	
	var lines []string
	for key, value := range colors {
		lines = append(lines, key+": "+value)
	}
	c.BrandColors = strings.Join(lines, "\n")
}

// ChatSession represents a chat session between a user and the assistant
type ChatSession struct {
	ID         string     `json:"id" db:"id"`
	CustomerID string     `json:"customer_id" db:"customer_id"`
	UserID     string     `json:"user_id" db:"user_id"`
	StartedAt  time.Time  `json:"started_at" db:"started_at"`
	EndedAt    *time.Time `json:"ended_at" db:"ended_at"`
	Metadata   string     `json:"metadata" db:"metadata"` // JSON string
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	ID        string    `json:"id" db:"id"`
	SessionID string    `json:"session_id" db:"session_id"`
	Role      string    `json:"role" db:"role"` // "user" or "assistant"
	Content   string    `json:"content" db:"content"`
	Emotion   string    `json:"emotion" db:"emotion"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AdminUser represents an administrator who can access the dashboard
type AdminUser struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// WidgetConfig represents the configuration sent to the widget
type WidgetConfig struct {
	CustomerID  string `json:"customerID"`
	BrandColors string `json:"brandColors"`
	LogoURL     string `json:"logoURL"`
	ModelPath   string `json:"modelPath"`
	Animations  string `json:"animations"`
	WSEndpoint  string `json:"wsEndpoint"`
}

// ChatMessageRequest represents an incoming chat message from the widget
type ChatMessageRequest struct {
	Content string `json:"content"`
	UserID  string `json:"userID"`
}

// ChatMessageResponse represents a response sent to the widget
type ChatMessageResponse struct {
	Content   string `json:"content"`
	Emotion   string `json:"emotion"`
	MessageID string `json:"messageId"`
}
