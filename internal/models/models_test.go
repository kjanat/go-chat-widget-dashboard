package models

import (
	"testing"
	"time"
)

func TestCustomerModel(t *testing.T) {
	customer := Customer{
		ID:             "test-id",
		Name:           "Test Customer",
		Email:          "test@example.com",
		CreatedAt:      time.Now(),
		BrandColors:    `{"primary": "#007bff"}`,
		LogoURL:        "https://example.com/logo.png",
		ModelPath:      "/models/test.glb",
		Animations:     `{"happy": {"name": "smile"}}`,
		OpenAIPrompt:   "You are a helpful assistant",
		AllowedDomains: "example.com, *.example.com",
		APIKey:         "test-api-key",
		Active:         true,
	}

	// Test that all fields are properly set
	if customer.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", customer.ID)
	}

	if customer.Name != "Test Customer" {
		t.Errorf("Expected Name to be 'Test Customer', got %s", customer.Name)
	}

	if !customer.Active {
		t.Error("Expected customer to be active")
	}
}

func TestChatMessageRequest(t *testing.T) {
	req := ChatMessageRequest{
		Content: "Hello, world!",
		UserID:  "user-123",
	}

	if req.Content != "Hello, world!" {
		t.Errorf("Expected content to be 'Hello, world!', got %s", req.Content)
	}

	if req.UserID != "user-123" {
		t.Errorf("Expected UserID to be 'user-123', got %s", req.UserID)
	}
}

func TestChatMessageResponse(t *testing.T) {
	resp := ChatMessageResponse{
		Content:   "Hi there!",
		Emotion:   "happy",
		MessageID: "msg-456",
	}

	if resp.Content != "Hi there!" {
		t.Errorf("Expected content to be 'Hi there!', got %s", resp.Content)
	}

	if resp.Emotion != "happy" {
		t.Errorf("Expected emotion to be 'happy', got %s", resp.Emotion)
	}
}
