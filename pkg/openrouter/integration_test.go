package openrouter

import (
	"context"
	"testing"
	"time"
)

// TestClientIntegration tests basic client functionality without making real API calls
func TestClientIntegration(t *testing.T) {
	// Test client creation and configuration
	client := NewClient("test-api-key")
	if client == nil {
		t.Fatal("Failed to create client")
	}

	// Test setting site info
	client.SetSiteInfo("https://example.com", "Test Bot")

	// Test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test building a request (this doesn't make an actual API call)
	req := ChatCompletionRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
		Temperature: func() *float32 { f := float32(0.7); return &f }(),
		MaxTokens:   func() *int { i := 100; return &i }(),
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		t.Errorf("Request validation failed: %v", err)
	}

	// Test building HTTP request
	httpReq, err := client.buildRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		t.Errorf("Failed to build request: %v", err)
	}

	// Verify headers are set correctly
	expectedHeaders := map[string]string{
		"Authorization": "Bearer test-api-key",
		"Content-Type":  "application/json",
		"HTTP-Referer":  "https://example.com",
		"X-Title":       "Test Bot",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := httpReq.Header.Get(key)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Verify URL is correct
	expectedURL := "https://openrouter.ai/api/v1/chat/completions"
	if httpReq.URL.String() != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, httpReq.URL.String())
	}
}

