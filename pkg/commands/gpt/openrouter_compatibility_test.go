package gpt

import (
	"testing"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
)

// TestOpenRouterMessageCompatibility tests that OpenRouter message types work with the cache
func TestOpenRouterMessageCompatibility(t *testing.T) {
	// Test that we can create cache data with OpenRouter types
	cacheData := &MessagesCacheData{
		Messages: []openrouter.ChatCompletionMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		SystemMessage: &openrouter.ChatCompletionMessage{
			Role: "system", Content: "You are a helpful assistant",
		},
		Model:      "openai/gpt-4",
		TokenCount: 50,
	}
	
	if cacheData == nil {
		t.Fatal("Cache data should not be nil")
	}
	
	if len(cacheData.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(cacheData.Messages))
	}
	
	if cacheData.SystemMessage == nil {
		t.Error("System message should not be nil")
	}
	
	if cacheData.Model != "openai/gpt-4" {
		t.Errorf("Expected model 'openai/gpt-4', got '%s'", cacheData.Model)
	}
}

// TestExtractBaseModelFunction tests the base model extraction for OpenRouter models
func TestExtractBaseModelFunction(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"openai/gpt-4", "gpt-4"},
		{"anthropic/claude-3-sonnet", "claude-3-sonnet"},
		{"gpt-3.5-turbo", "gpt-3.5-turbo"},
		{"", ""},
	}
	
	for _, tc := range testCases {
		result := extractBaseModel(tc.input)
		if result != tc.expected {
			t.Errorf("extractBaseModel(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TestTokenCountingWithOpenRouterModels tests that token counting works with OpenRouter models
func TestTokenCountingWithOpenRouterModels(t *testing.T) {
	message := openrouter.ChatCompletionMessage{
		Role:    "user",
		Content: "Hello, how are you?",
	}
	
	// Test with OpenRouter model format
	result := countOpenRouterMessageTokens(message, "openai/gpt-4")
	if result == nil {
		t.Error("Token count should not be nil for valid message")
	}
	
	if result != nil && *result <= 0 {
		t.Errorf("Token count should be positive, got %d", *result)
	}
}