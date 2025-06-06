package services

import (
	"testing"

	"github.com/kjanat/go-chat-widget-dashboard/internal/database"
)

func TestOpenAIService_GenerateResponse(t *testing.T) {
	service := NewOpenAIService()

	prompt := "You are a helpful assistant"
	message := "Hello, how are you?"

	response, emotion := service.GenerateResponse(prompt, message)

	// Test that we get a non-empty response
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Test that emotion is one of the expected values
	validEmotions := map[string]bool{
		"happy":    true,
		"thinking": true,
		"neutral":  true,
		"excited":  true,
	}

	if !validEmotions[emotion] {
		t.Errorf("Expected valid emotion, got %s", emotion)
	}

	// Test that response contains our message
	if len(response) < len(message) {
		t.Error("Response should be longer than input message")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	// IDs should not be empty
	if id1 == "" || id2 == "" {
		t.Error("Generated IDs should not be empty")
	}

	// IDs should be different (with very high probability)
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

// helper to setup DB and customer service with configurable allowed domains
func setupCustomerService(t *testing.T, allowed string) (*CustomerService, *database.DB, string) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	customerID := "c1"
	_, err = db.Exec(`INSERT INTO customers (id, name, email, created_at, brand_colors, logo_url, model_path, animations, openai_prompt, allowed_domains, api_key, active) VALUES (?, 'Test', 'test@example.com', datetime('now'), '', '', '', '', '', ?, 'key', 1)`,
		customerID, allowed)
	if err != nil {
		t.Fatalf("failed to insert customer: %v", err)
	}

	return NewCustomerService(db.DB), db, customerID
}

func TestValidateOriginWildcard(t *testing.T) {
	svc, db, id := setupCustomerService(t, "*")
	defer db.Close()

	if !svc.ValidateOrigin("https://example.com", id) {
		t.Error("expected origin to be allowed when wildcard is set")
	}
}

func TestValidateOriginEmpty(t *testing.T) {
	svc, db, id := setupCustomerService(t, "")
	defer db.Close()

	if svc.ValidateOrigin("https://example.com", id) {
		t.Error("expected origin to be rejected when allowed_domains is empty")
	}
}
