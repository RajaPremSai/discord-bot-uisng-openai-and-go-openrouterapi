package gpt

import (
	"testing"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
)

// TestNormalizeOpenRouterModelName tests the model name normalization function
func TestNormalizeOpenRouterModelName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"OpenAI GPT-4", "openai/gpt-4", "GPT-4"},
		{"OpenAI GPT-3.5", "openai/gpt-3.5-turbo", "GPT-3.5-TURBO"},
		{"Anthropic Claude", "anthropic/claude-3-sonnet", "Claude-3-Sonnet"},
		{"Anthropic Claude Haiku", "anthropic/claude-3-haiku", "Claude-3-Haiku"},
		{"Direct model GPT", "gpt-4", "gpt-4"},
		{"Direct model Claude", "claude-3-sonnet", "claude-3-sonnet"},
		{"Empty model", "", ""},
		{"Model without provider", "some-model", "some-model"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeOpenRouterModelName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeOpenRouterModelName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestOpenRouterErrorHandling tests that OpenRouter errors are properly handled
func TestOpenRouterErrorHandling(t *testing.T) {
	testCases := []struct {
		name      string
		errorCode string
		model     string
	}{
		{"Insufficient credits", "insufficient_quota", "openai/gpt-4"},
		{"Model not found", "model_not_found", "openai/gpt-5"},
		{"Rate limit exceeded", "rate_limit_exceeded", "anthropic/claude-3-sonnet"},
		{"Invalid request", "invalid_request_error", "openai/gpt-4"},
		{"Context length exceeded", "context_length_exceeded", "openai/gpt-3.5-turbo"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock OpenRouter error
			err := &openrouter.ErrorResponse{
				ErrorDetail: openrouter.ErrorDetail{
					Code:    tc.errorCode,
					Message: "Test error message",
					Type:    "error",
				},
			}

			// Test that the error implements the error interface
			if err.Error() == "" {
				t.Error("OpenRouter error should implement error interface")
			}

			// Test that error code is accessible
			if err.ErrorDetail.Code != tc.errorCode {
				t.Errorf("Expected error code %s, got %s", tc.errorCode, err.ErrorDetail.Code)
			}
		})
	}
}

// TestMessagesCacheDataWithOpenRouterTypes tests cache data with various OpenRouter message types
func TestMessagesCacheDataWithOpenRouterTypes(t *testing.T) {
	// Test with system message
	cacheData := &MessagesCacheData{
		Messages: []openrouter.ChatCompletionMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
		SystemMessage: &openrouter.ChatCompletionMessage{
			Role:    "system",
			Content: "You are a helpful assistant specialized in programming.",
		},
		Model:       "openai/gpt-4",
		Temperature: func() *float32 { t := float32(0.7); return &t }(),
		TokenCount:  150,
	}

	// Test validation
	if !cacheData.ValidateOpenRouterModel() {
		t.Error("Valid OpenRouter model should pass validation")
	}

	// Test normalized name
	normalizedName := cacheData.GetNormalizedModelName()
	if normalizedName != "GPT-4" {
		t.Errorf("Expected normalized name 'GPT-4', got '%s'", normalizedName)
	}

	// Test base model name
	baseName := cacheData.GetBaseModelName()
	if baseName != "gpt-4" {
		t.Errorf("Expected base name 'gpt-4', got '%s'", baseName)
	}

	// Test with different model formats
	testModels := []string{
		"anthropic/claude-3-sonnet",
		"openai/gpt-3.5-turbo",
		"gpt-4-turbo",
		"claude-3-haiku",
	}

	for _, model := range testModels {
		cacheData.Model = model
		if !cacheData.ValidateOpenRouterModel() {
			t.Errorf("Model %s should be valid", model)
		}
	}
}

// TestTokenCountingAccuracy tests that token counting is reasonably accurate for OpenRouter messages
func TestTokenCountingAccuracy(t *testing.T) {
	testCases := []struct {
		name     string
		message  openrouter.ChatCompletionMessage
		model    string
		minTokens int
		maxTokens int
	}{
		{
			name:      "Short message",
			message:   openrouter.ChatCompletionMessage{Role: "user", Content: "Hello"},
			model:     "openai/gpt-4",
			minTokens: 1,
			maxTokens: 10,
		},
		{
			name:      "Medium message",
			message:   openrouter.ChatCompletionMessage{Role: "user", Content: "This is a longer message with multiple words and punctuation."},
			model:     "openai/gpt-3.5-turbo",
			minTokens: 10,
			maxTokens: 25,
		},
		{
			name:      "Long message",
			message:   openrouter.ChatCompletionMessage{Role: "assistant", Content: "This is a very long message that contains multiple sentences and should result in a higher token count. It includes various punctuation marks, numbers like 123, and different types of content that would be typical in a conversation."},
			model:     "anthropic/claude-3-sonnet",
			minTokens: 30,
			maxTokens: 80,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenCount := countOpenRouterMessageTokens(tc.message, tc.model)
			if tokenCount == nil {
				t.Error("Token count should not be nil")
				return
			}

			if *tokenCount < tc.minTokens || *tokenCount > tc.maxTokens {
				t.Errorf("Token count %d is outside expected range [%d, %d] for message: %s", 
					*tokenCount, tc.minTokens, tc.maxTokens, tc.message.Content)
			}
		})
	}
}

// TestMessagesCacheIntegration tests the integration of caching with OpenRouter message handling
func TestMessagesCacheIntegration(t *testing.T) {
	cache, err := NewMessagesCache(5)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test multiple channels with different models
	channels := []struct {
		id    string
		model string
	}{
		{"channel1", "openai/gpt-4"},
		{"channel2", "anthropic/claude-3-sonnet"},
		{"channel3", "openai/gpt-3.5-turbo"},
	}

	for _, ch := range channels {
		cacheData := &MessagesCacheData{
			Messages: []openrouter.ChatCompletionMessage{
				{Role: "user", Content: "Test message for " + ch.id},
			},
			Model:      ch.model,
			TokenCount: 10,
		}

		cache.Add(ch.id, cacheData)
	}

	// Verify all channels are cached
	for _, ch := range channels {
		retrieved, ok := cache.Get(ch.id)
		if !ok {
			t.Errorf("Failed to retrieve cache for channel %s", ch.id)
			continue
		}

		if retrieved.Model != ch.model {
			t.Errorf("Expected model %s for channel %s, got %s", ch.model, ch.id, retrieved.Model)
		}

		if len(retrieved.Messages) != 1 {
			t.Errorf("Expected 1 message for channel %s, got %d", ch.id, len(retrieved.Messages))
		}
	}

	// Test cache eviction (cache size is 5, so adding more should evict older entries)
	for i := 0; i < 10; i++ {
		channelID := "temp_channel_" + string(rune('0'+i))
		cacheData := &MessagesCacheData{
			Messages: []openrouter.ChatCompletionMessage{
				{Role: "user", Content: "Temp message"},
			},
			Model:      "openai/gpt-4",
			TokenCount: 5,
		}
		cache.Add(channelID, cacheData)
	}

	// Some of the original channels should be evicted
	evictedCount := 0
	for _, ch := range channels {
		if _, ok := cache.Get(ch.id); !ok {
			evictedCount++
		}
	}

	if evictedCount == 0 {
		t.Error("Expected some channels to be evicted from cache")
	}
}