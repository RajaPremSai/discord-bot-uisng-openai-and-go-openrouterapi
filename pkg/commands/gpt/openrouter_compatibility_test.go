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

// TestMessagesCacheDataValidation tests the validation methods for cache data
func TestMessagesCacheDataValidation(t *testing.T) {
	testCases := []struct {
		name     string
		model    string
		expected bool
	}{
		{"Valid OpenRouter format", "openai/gpt-4", true},
		{"Valid Anthropic format", "anthropic/claude-3-sonnet", true},
		{"Valid direct model", "gpt-4", true},
		{"Empty model", "", false},
		{"Invalid format with slash", "openai/", false},
		{"Invalid format with multiple slashes", "openai/gpt/4", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cacheData := &MessagesCacheData{
				Model: tc.model,
			}
			
			result := cacheData.ValidateOpenRouterModel()
			if result != tc.expected {
				t.Errorf("ValidateOpenRouterModel() = %v, want %v for model %s", result, tc.expected, tc.model)
			}
		})
	}
}

// TestMessagesCacheDataNormalization tests model name normalization
func TestMessagesCacheDataNormalization(t *testing.T) {
	testCases := []struct {
		name     string
		model    string
		expected string
	}{
		{"OpenAI GPT-4", "openai/gpt-4", "GPT-4"},
		{"Anthropic Claude", "anthropic/claude-3-sonnet", "Claude-3-Sonnet"},
		{"Direct model", "gpt-3.5-turbo", "gpt-3.5-turbo"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cacheData := &MessagesCacheData{
				Model: tc.model,
			}
			
			result := cacheData.GetNormalizedModelName()
			if result != tc.expected {
				t.Errorf("GetNormalizedModelName() = %q, want %q for model %s", result, tc.expected, tc.model)
			}
		})
	}
}

// TestMessagesCacheDataBaseModel tests base model extraction
func TestMessagesCacheDataBaseModel(t *testing.T) {
	testCases := []struct {
		name     string
		model    string
		expected string
	}{
		{"OpenAI GPT-4", "openai/gpt-4", "gpt-4"},
		{"Anthropic Claude", "anthropic/claude-3-sonnet", "claude-3-sonnet"},
		{"Direct model", "gpt-3.5-turbo", "gpt-3.5-turbo"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cacheData := &MessagesCacheData{
				Model: tc.model,
			}
			
			result := cacheData.GetBaseModelName()
			if result != tc.expected {
				t.Errorf("GetBaseModelName() = %q, want %q for model %s", result, tc.expected, tc.model)
			}
		})
	}
}

// TestOpenRouterMessageCaching tests message caching with OpenRouter data structures
func TestOpenRouterMessageCaching(t *testing.T) {
	cache, err := NewMessagesCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	
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
	
	// Test adding to cache
	cache.Add("test-channel", cacheData)
	
	// Test retrieving from cache
	retrieved, ok := cache.Get("test-channel")
	if !ok {
		t.Error("Failed to retrieve cache data")
	}
	
	if retrieved.Model != "openai/gpt-4" {
		t.Errorf("Expected model 'openai/gpt-4', got '%s'", retrieved.Model)
	}
	
	if len(retrieved.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved.Messages))
	}
	
	if retrieved.TokenCount != 50 {
		t.Errorf("Expected token count 50, got %d", retrieved.TokenCount)
	}
}

// TestOpenRouterTokenTruncation tests token counting and truncation with OpenRouter models
func TestOpenRouterTokenTruncation(t *testing.T) {
	cacheData := &MessagesCacheData{
		Messages: []openrouter.ChatCompletionMessage{
			{Role: "user", Content: "This is a test message"},
			{Role: "assistant", Content: "This is a response"},
			{Role: "user", Content: "Another message"},
		},
		Model:      "openai/gpt-4",
		TokenCount: 0,
	}
	
	// Test token counting
	ok, count := isCacheItemWithinTruncateLimit(cacheData)
	if !ok && count <= 0 {
		t.Error("Token counting should work for valid messages")
	}
	
	// Verify token count was updated
	if cacheData.TokenCount <= 0 {
		t.Error("Token count should be updated after checking limits")
	}
}

// TestOpenRouterModelTruncateLimits tests truncate limits for different OpenRouter models
func TestOpenRouterModelTruncateLimits(t *testing.T) {
	testCases := []struct {
		name     string
		model    string
		hasLimit bool
	}{
		{"OpenAI GPT-4", "openai/gpt-4", true},
		{"OpenAI GPT-3.5", "openai/gpt-3.5-turbo", true},
		{"Anthropic Claude", "anthropic/claude-3-sonnet", true},
		{"Direct GPT-4", "gpt-4", true},
		{"GPT-4 Turbo", "openai/gpt-4-turbo", true},
		{"GPT-4 32k", "openai/gpt-4-32k", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limit := modelTruncateLimit(tc.model)
			
			if tc.hasLimit && limit == nil {
				t.Errorf("Expected truncate limit for model %s, got nil", tc.model)
			}
			
			if tc.hasLimit && limit != nil && *limit <= 0 {
				t.Errorf("Expected positive truncate limit for model %s, got %d", tc.model, *limit)
			}
		})
	}
}