package services

import (
	"testing"
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
