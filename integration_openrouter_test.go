package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test configuration
const (
	testTimeout = 30 * time.Second
	testModel   = "openai/gpt-3.5-turbo"
	testImageModel = "openai/dall-e-2"
)

// getTestAPIKey retrieves the OpenRouter API key from environment or credentials file
func getTestAPIKey(t *testing.T) string {
	// First try environment variable
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		return apiKey
	}

	// Try to read from credentials file
	config := &Config{}
	if err := config.ReadFromFile("credentials.yaml"); err == nil {
		if config.OpenRouter.APIKey != "" && config.OpenRouter.APIKey != "sk-or-v1-your-api-key-here" {
			return config.OpenRouter.APIKey
		}
	}

	t.Skip("OpenRouter API key not found. Set OPENROUTER_API_KEY environment variable or configure credentials.yaml")
	return ""
}

// createTestClient creates a test OpenRouter client with proper configuration
func createTestClient(t *testing.T) *openrouter.Client {
	apiKey := getTestAPIKey(t)
	
	client := openrouter.NewClientWithConfig(openrouter.ClientConfig{
		APIKey:   apiKey,
		BaseURL:  "https://openrouter.ai/api/v1",
		SiteURL:  "https://test-integration.com",
		SiteName: "OpenRouter Integration Test",
		Logger:   openrouter.DefaultLogger(),
	})

	return client
}

// TestOpenRouterChatCompletionIntegration tests chat completion with real OpenRouter API calls
func TestOpenRouterChatCompletionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := createTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("BasicChatCompletion", func(t *testing.T) {
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Hello! Please respond with exactly 'Integration test successful'",
				},
			},
			MaxTokens:   openrouter.IntPtr(50),
			Temperature: openrouter.Float32Ptr(0.1),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Chat completion should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Choices, "Response should have choices")
		
		choice := resp.Choices[0]
		assert.NotEmpty(t, choice.Message.Content, "Response content should not be empty")
		assert.Equal(t, "assistant", choice.Message.Role, "Response role should be assistant")
		
		// Verify usage information
		require.NotNil(t, resp.Usage, "Usage information should be present")
		assert.Greater(t, resp.Usage.PromptTokens, 0, "Prompt tokens should be greater than 0")
		assert.Greater(t, resp.Usage.CompletionTokens, 0, "Completion tokens should be greater than 0")
		assert.Greater(t, resp.Usage.TotalTokens, 0, "Total tokens should be greater than 0")
		assert.Equal(t, resp.Usage.PromptTokens+resp.Usage.CompletionTokens, resp.Usage.TotalTokens, "Total tokens should equal prompt + completion tokens")

		t.Logf("Chat completion successful. Model: %s, Usage: %+v", resp.Model, resp.Usage)
	})

	t.Run("ChatCompletionWithSystemMessage", func(t *testing.T) {
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a helpful assistant that always responds in exactly 5 words.",
				},
				{
					Role:    "user",
					Content: "What is the capital of France?",
				},
			},
			MaxTokens:   openrouter.IntPtr(20),
			Temperature: openrouter.Float32Ptr(0.3),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Chat completion with system message should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Choices, "Response should have choices")
		
		choice := resp.Choices[0]
		assert.NotEmpty(t, choice.Message.Content, "Response content should not be empty")
		
		// Count words in response (should be approximately 5 due to system message)
		words := strings.Fields(choice.Message.Content)
		assert.LessOrEqual(t, len(words), 10, "Response should be relatively short due to system message constraint")

		t.Logf("System message chat completion successful. Response: %s", choice.Message.Content)
	})

	t.Run("ChatCompletionWithMultipleMessages", func(t *testing.T) {
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "My name is Alice.",
				},
				{
					Role:    "assistant",
					Content: "Hello Alice! Nice to meet you.",
				},
				{
					Role:    "user",
					Content: "What is my name?",
				},
			},
			MaxTokens:   openrouter.IntPtr(30),
			Temperature: openrouter.Float32Ptr(0.1),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Multi-message chat completion should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Choices, "Response should have choices")
		
		choice := resp.Choices[0]
		assert.NotEmpty(t, choice.Message.Content, "Response content should not be empty")
		
		// Response should mention Alice since the conversation context includes the name
		responseContent := strings.ToLower(choice.Message.Content)
		assert.Contains(t, responseContent, "alice", "Response should remember the user's name from conversation context")

		t.Logf("Multi-message chat completion successful. Response: %s", choice.Message.Content)
	})
}

// TestOpenRouterImageGenerationIntegration tests image generation with real OpenRouter API calls
func TestOpenRouterImageGenerationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := createTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("BasicImageGeneration", func(t *testing.T) {
		req := openrouter.ImageRequest{
			Prompt:         "A simple red circle on a white background",
			Model:          testImageModel,
			N:              1,
			Size:           "256x256",
			ResponseFormat: "url",
		}

		resp, err := client.CreateImage(ctx, req)
		require.NoError(t, err, "Image generation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Data, "Response should have image data")
		require.Len(t, resp.Data, 1, "Should generate exactly 1 image")
		
		imageData := resp.Data[0]
		assert.NotEmpty(t, imageData.URL, "Image URL should not be empty")
		assert.True(t, strings.HasPrefix(imageData.URL, "http"), "Image URL should be a valid HTTP URL")

		t.Logf("Image generation successful. URL: %s", imageData.URL)
	})

	t.Run("MultipleImageGeneration", func(t *testing.T) {
		req := openrouter.ImageRequest{
			Prompt:         "A cute cartoon cat",
			Model:          testImageModel,
			N:              2,
			Size:           "256x256",
			ResponseFormat: "url",
		}

		resp, err := client.CreateImage(ctx, req)
		require.NoError(t, err, "Multiple image generation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Data, "Response should have image data")
		require.Len(t, resp.Data, 2, "Should generate exactly 2 images")
		
		for i, imageData := range resp.Data {
			assert.NotEmpty(t, imageData.URL, "Image URL %d should not be empty", i+1)
			assert.True(t, strings.HasPrefix(imageData.URL, "http"), "Image URL %d should be a valid HTTP URL", i+1)
		}

		t.Logf("Multiple image generation successful. Generated %d images", len(resp.Data))
	})

	t.Run("LargerImageGeneration", func(t *testing.T) {
		req := openrouter.ImageRequest{
			Prompt:         "A beautiful sunset over mountains",
			Model:          testImageModel,
			N:              1,
			Size:           "512x512",
			ResponseFormat: "url",
		}

		resp, err := client.CreateImage(ctx, req)
		require.NoError(t, err, "Larger image generation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Data, "Response should have image data")
		require.Len(t, resp.Data, 1, "Should generate exactly 1 image")
		
		imageData := resp.Data[0]
		assert.NotEmpty(t, imageData.URL, "Image URL should not be empty")
		assert.True(t, strings.HasPrefix(imageData.URL, "http"), "Image URL should be a valid HTTP URL")

		t.Logf("Larger image generation successful. URL: %s", imageData.URL)
	})
}

// TestOpenRouterErrorScenariosIntegration tests error scenarios and rate limiting
func TestOpenRouterErrorScenariosIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InvalidAPIKey", func(t *testing.T) {
		client := openrouter.NewClient("sk-or-v1-invalid-key-12345")
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		}

		_, err := client.CreateChatCompletion(ctx, req)
		require.Error(t, err, "Should fail with invalid API key")
		
		if orErr, ok := err.(*openrouter.OpenRouterError); ok {
			assert.Equal(t, 401, orErr.StatusCode, "Should return 401 Unauthorized")
			assert.False(t, orErr.IsRetryable, "Authentication errors should not be retryable")
			assert.Contains(t, strings.ToLower(orErr.GetUserMessage()), "authentication", "Error message should mention authentication")
		}

		t.Logf("Invalid API key error handled correctly: %v", err)
	})

	t.Run("InvalidModel", func(t *testing.T) {
		client := createTestClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		req := openrouter.ChatCompletionRequest{
			Model: "invalid/nonexistent-model",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		}

		_, err := client.CreateChatCompletion(ctx, req)
		require.Error(t, err, "Should fail with invalid model")
		
		if orErr, ok := err.(*openrouter.OpenRouterError); ok {
			// Could be 400 (bad request) or 404 (not found) depending on OpenRouter's response
			assert.True(t, orErr.StatusCode == 400 || orErr.StatusCode == 404, "Should return 400 or 404 for invalid model")
			assert.False(t, orErr.IsRetryable, "Invalid model errors should not be retryable")
		}

		t.Logf("Invalid model error handled correctly: %v", err)
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		client := createTestClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "",
				},
			},
		}

		_, err := client.CreateChatCompletion(ctx, req)
		require.Error(t, err, "Should fail with empty prompt")
		
		// This could fail at validation level or API level
		t.Logf("Empty prompt error handled correctly: %v", err)
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		client := createTestClient(t)
		// Very short timeout to force timeout error
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		}

		_, err := client.CreateChatCompletion(ctx, req)
		require.Error(t, err, "Should fail with context timeout")
		
		// Should be context deadline exceeded
		assert.Contains(t, err.Error(), "context deadline exceeded", "Error should mention context deadline")

		t.Logf("Context timeout error handled correctly: %v", err)
	})
}

// TestOpenRouterRetryLogicIntegration tests retry logic with real scenarios
func TestOpenRouterRetryLogicIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := createTestClient(t)

	t.Run("RetryWithValidRequest", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		retryConfig := &openrouter.RetryConfig{
			MaxRetries:    2,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffFactor: 2.0,
			JitterEnabled: false,
		}

		var attempts int
		err := client.WithRetry(ctx, retryConfig, func() error {
			attempts++
			
			// Simulate success on second attempt
			if attempts == 1 {
				// Return a retryable error for first attempt
				return &openrouter.OpenRouterError{
					StatusCode:  503,
					ErrorCode:   "service_unavailable",
					Message:     "Service temporarily unavailable",
					IsRetryable: true,
				}
			}
			
			// Success on second attempt
			return nil
		})

		require.NoError(t, err, "Retry should eventually succeed")
		assert.Equal(t, 2, attempts, "Should make exactly 2 attempts")

		t.Logf("Retry logic test successful after %d attempts", attempts)
	})

	t.Run("RetryExhaustion", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		retryConfig := &openrouter.RetryConfig{
			MaxRetries:    2,
			BaseDelay:     50 * time.Millisecond,
			MaxDelay:      500 * time.Millisecond,
			BackoffFactor: 2.0,
			JitterEnabled: false,
		}

		var attempts int
		err := client.WithRetry(ctx, retryConfig, func() error {
			attempts++
			// Always return retryable error
			return &openrouter.OpenRouterError{
				StatusCode:  503,
				ErrorCode:   "service_unavailable",
				Message:     "Service temporarily unavailable",
				IsRetryable: true,
			}
		})

		require.Error(t, err, "Should fail after exhausting retries")
		assert.Equal(t, 3, attempts, "Should make 3 attempts (initial + 2 retries)")

		t.Logf("Retry exhaustion test successful after %d attempts", attempts)
	})
}

// TestOpenRouterConnectionAndModels tests connection and model listing
func TestOpenRouterConnectionAndModels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := createTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("PingConnection", func(t *testing.T) {
		err := client.Ping(ctx)
		require.NoError(t, err, "Ping should succeed with valid API key")

		t.Log("OpenRouter API connection test successful")
	})

	t.Run("ListModels", func(t *testing.T) {
		resp, err := client.ListModels(ctx)
		require.NoError(t, err, "List models should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Data, "Should have at least some models")

		// Verify that our test models are available
		var foundChatModel, foundImageModel bool
		for _, model := range resp.Data {
			if model.ID == testModel {
				foundChatModel = true
			}
			if model.ID == testImageModel {
				foundImageModel = true
			}
		}

		assert.True(t, foundChatModel, "Test chat model %s should be available", testModel)
		assert.True(t, foundImageModel, "Test image model %s should be available", testImageModel)

		t.Logf("Found %d available models", len(resp.Data))
	})

	t.Run("GetSpecificModel", func(t *testing.T) {
		model, err := client.GetModel(ctx, testModel)
		require.NoError(t, err, "Get model should succeed")
		require.NotNil(t, model, "Model should not be nil")
		
		assert.Equal(t, testModel, model.ID, "Model ID should match requested model")
		assert.NotEmpty(t, model.ID, "Model ID should not be empty")

		t.Logf("Retrieved model info: ID=%s, Object=%s", model.ID, model.Object)
	})
}

// TestEndToEndDiscordCommandFlow tests the complete flow from Discord command to OpenRouter response
func TestEndToEndDiscordCommandFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test simulates the complete flow without actually using Discord
	// It tests the same code paths that would be used in a real Discord interaction

	client := createTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("SimulateChatCommand", func(t *testing.T) {
		// Simulate the request that would be created by the Discord chat command
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a helpful Discord bot assistant.",
				},
				{
					Role:    "user",
					Content: "Hello! Can you help me with a simple math problem? What is 2 + 2?",
				},
			},
			MaxTokens:   openrouter.IntPtr(100),
			Temperature: openrouter.Float32Ptr(0.7),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Chat command simulation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Choices, "Response should have choices")
		
		choice := resp.Choices[0]
		assert.NotEmpty(t, choice.Message.Content, "Response content should not be empty")
		assert.Equal(t, "assistant", choice.Message.Role, "Response role should be assistant")
		
		// The response should contain "4" since we asked for 2+2
		responseContent := strings.ToLower(choice.Message.Content)
		assert.Contains(t, responseContent, "4", "Response should contain the answer '4'")

		t.Logf("Chat command simulation successful. Response: %s", choice.Message.Content)
	})

	t.Run("SimulateImageCommand", func(t *testing.T) {
		// Simulate the request that would be created by the Discord image command
		req := openrouter.ImageRequest{
			Prompt:         "A friendly robot waving hello",
			Model:          testImageModel,
			N:              1,
			Size:           "256x256",
			ResponseFormat: "url",
		}

		resp, err := client.CreateImage(ctx, req)
		require.NoError(t, err, "Image command simulation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Data, "Response should have image data")
		require.Len(t, resp.Data, 1, "Should generate exactly 1 image")
		
		imageData := resp.Data[0]
		assert.NotEmpty(t, imageData.URL, "Image URL should not be empty")
		assert.True(t, strings.HasPrefix(imageData.URL, "http"), "Image URL should be a valid HTTP URL")

		t.Logf("Image command simulation successful. URL: %s", imageData.URL)
	})

	t.Run("SimulateChatWithContext", func(t *testing.T) {
		// Simulate a chat command with context file (system message)
		contextContent := "You are an expert in Go programming. Always provide concise, accurate answers about Go."
		
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "system",
					Content: contextContent,
				},
				{
					Role:    "user",
					Content: "What is a goroutine?",
				},
			},
			MaxTokens:   openrouter.IntPtr(150),
			Temperature: openrouter.Float32Ptr(0.3),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Chat with context simulation should succeed")
		require.NotNil(t, resp, "Response should not be nil")
		require.NotEmpty(t, resp.Choices, "Response should have choices")
		
		choice := resp.Choices[0]
		assert.NotEmpty(t, choice.Message.Content, "Response content should not be empty")
		
		// Response should mention goroutines since that's what we asked about
		responseContent := strings.ToLower(choice.Message.Content)
		assert.True(t, 
			strings.Contains(responseContent, "goroutine") || 
			strings.Contains(responseContent, "concurrent") || 
			strings.Contains(responseContent, "go"), 
			"Response should be relevant to Go programming and goroutines")

		t.Logf("Chat with context simulation successful. Response: %s", choice.Message.Content)
	})
}

// TestOpenRouterLoggingIntegration tests that logging works correctly during API calls
func TestOpenRouterLoggingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create client with custom logger configuration
	logger := openrouter.NewLogger(openrouter.LoggerConfig{
		Level:             openrouter.LogLevelDebug,
		EnableMetrics:     true,
		EnableRequestLog:  true,
		EnableResponseLog: true,
	})

	client := openrouter.NewClientWithConfig(openrouter.ClientConfig{
		APIKey:   getTestAPIKey(t),
		BaseURL:  "https://openrouter.ai/api/v1",
		SiteURL:  "https://test-logging.com",
		SiteName: "Logging Test",
		Logger:   logger,
	})

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("LoggingDuringChatCompletion", func(t *testing.T) {
		req := openrouter.ChatCompletionRequest{
			Model: testModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Test logging functionality",
				},
			},
			MaxTokens:   openrouter.IntPtr(20),
			Temperature: openrouter.Float32Ptr(0.5),
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		require.NoError(t, err, "Chat completion should succeed")
		require.NotNil(t, resp, "Response should not be nil")

		// The logging happens internally, we just verify the request succeeded
		// In a real scenario, you would capture log output to verify logging
		t.Log("Chat completion with logging successful")
	})

	t.Run("LoggingDuringError", func(t *testing.T) {
		// Use invalid model to trigger error logging
		req := openrouter.ChatCompletionRequest{
			Model: "invalid/model-for-logging-test",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "This should fail",
				},
			},
		}

		_, err := client.CreateChatCompletion(ctx, req)
		require.Error(t, err, "Should fail with invalid model")

		// Error logging happens internally
		t.Log("Error logging test completed")
	})
}