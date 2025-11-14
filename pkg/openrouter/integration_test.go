package openrouter

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Integration test configuration
const (
	testAPIKeyEnv     = "OPENROUTER_API_KEY"
	testTimeout       = 30 * time.Second
	testModel         = "openai/gpt-3.5-turbo"
	testImageModel    = "openai/dall-e-2"
	testPrompt        = "Hello, this is a test message for integration testing."
	testImagePrompt   = "A simple test image of a red circle"
	testSiteURL       = "https://test-integration.com"
	testSiteName      = "OpenRouter Integration Test"
)

// skipIfNoAPIKey skips the test if no API key is provided
func skipIfNoAPIKey(t *testing.T) string {
	apiKey := os.Getenv(testAPIKeyEnv)
	if apiKey == "" {
		t.Skipf("Skipping integration test: %s environment variable not set", testAPIKeyEnv)
	}
	return apiKey
}

// createTestClient creates a client for integration testing
func createTestClient(t *testing.T) *Client {
	apiKey := skipIfNoAPIKey(t)
	
	config := ClientConfig{
		APIKey:   apiKey,
		BaseURL:  DefaultBaseURL,
		SiteURL:  testSiteURL,
		SiteName: testSiteName,
		Logger:   DefaultLogger(),
	}
	
	client := NewClientWithConfig(config)
	return client
}

// TestIntegration_ChatCompletion tests chat completion with real OpenRouter API calls
func TestIntegration_ChatCompletion(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("basic_chat_completion", func(t *testing.T) {
		req := ChatCompletionRequest{
			Model: testModel,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}
		
		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			t.Fatalf("CreateChatCompletion failed: %v", err)
		}
		
		// Validate response structure
		if resp.ID == "" {
			t.Error("Response ID is empty")
		}
		if resp.Object != "chat.completion" {
			t.Errorf("Expected object 'chat.completion', got '%s'", resp.Object)
		}
		if resp.Model == "" {
			t.Error("Response model is empty")
		}
		if len(resp.Choices) == 0 {
			t.Error("No choices in response")
		}
		if resp.Choices[0].Message.Content == "" {
			t.Error("Response content is empty")
		}
		if resp.Usage == nil {
			t.Error("Usage information is missing")
		}
		if resp.Usage.TotalTokens == 0 {
			t.Error("Total tokens should be greater than 0")
		}
		
		t.Logf("Chat completion successful: %d tokens used", resp.Usage.TotalTokens)
	})
	
	t.Run("chat_completion_with_temperature", func(t *testing.T) {
		temperature := float32(0.7)
		req := ChatCompletionRequest{
			Model:       testModel,
			Temperature: &temperature,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}
		
		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			t.Fatalf("CreateChatCompletion with temperature failed: %v", err)
		}
		
		if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
			t.Error("Invalid response with temperature parameter")
		}
		
		t.Logf("Chat completion with temperature successful")
	})
	
	t.Run("chat_completion_with_max_tokens", func(t *testing.T) {
		maxTokens := 50
		req := ChatCompletionRequest{
			Model:     testModel,
			MaxTokens: &maxTokens,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Write a very long story about a dragon.",
				},
			},
		}
		
		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			t.Fatalf("CreateChatCompletion with max tokens failed: %v", err)
		}
		
		if resp.Usage.CompletionTokens > maxTokens {
			t.Errorf("Response exceeded max tokens: got %d, max %d", resp.Usage.CompletionTokens, maxTokens)
		}
		
		t.Logf("Chat completion with max tokens successful: %d tokens used", resp.Usage.CompletionTokens)
	})
	
	t.Run("chat_completion_with_system_message", func(t *testing.T) {
		req := ChatCompletionRequest{
			Model: testModel,
			Messages: []ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a helpful assistant that responds in exactly 5 words.",
				},
				{
					Role:    "user",
					Content: "What is the weather like?",
				},
			},
		}
		
		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			t.Fatalf("CreateChatCompletion with system message failed: %v", err)
		}
		
		if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
			t.Error("Invalid response with system message")
		}
		
		// Check if response is roughly 5 words (allowing some flexibility)
		words := strings.Fields(resp.Choices[0].Message.Content)
		if len(words) > 10 {
			t.Logf("Warning: Response has %d words, expected around 5: %s", len(words), resp.Choices[0].Message.Content)
		}
		
		t.Logf("Chat completion with system message successful")
	})
}

// TestIntegration_ImageGeneration tests image generation with real OpenRouter API calls
func TestIntegration_ImageGeneration(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("basic_image_generation", func(t *testing.T) {
		req := ImageRequest{
			Prompt: testImagePrompt,
			Model:  testImageModel,
			N:      1,
			Size:   "256x256",
		}
		
		resp, err := client.CreateImage(ctx, req)
		if err != nil {
			t.Fatalf("CreateImage failed: %v", err)
		}
		
		// Validate response structure
		if len(resp.Data) == 0 {
			t.Error("No image data in response")
		}
		if resp.Data[0].URL == "" {
			t.Error("Image URL is empty")
		}
		if !strings.HasPrefix(resp.Data[0].URL, "http") {
			t.Errorf("Invalid image URL format: %s", resp.Data[0].URL)
		}
		
		t.Logf("Image generation successful: %s", resp.Data[0].URL)
	})
	
	t.Run("multiple_images", func(t *testing.T) {
		req := ImageRequest{
			Prompt: testImagePrompt,
			Model:  testImageModel,
			N:      2,
			Size:   "256x256",
		}
		
		resp, err := client.CreateImage(ctx, req)
		if err != nil {
			t.Fatalf("CreateImage with multiple images failed: %v", err)
		}
		
		if len(resp.Data) != 2 {
			t.Errorf("Expected 2 images, got %d", len(resp.Data))
		}
		
		for i, img := range resp.Data {
			if img.URL == "" {
				t.Errorf("Image %d URL is empty", i)
			}
		}
		
		t.Logf("Multiple image generation successful: %d images", len(resp.Data))
	})
	
	t.Run("different_image_sizes", func(t *testing.T) {
		sizes := []string{"256x256", "512x512"}
		
		for _, size := range sizes {
			t.Run("size_"+size, func(t *testing.T) {
				req := ImageRequest{
					Prompt: testImagePrompt,
					Model:  testImageModel,
					N:      1,
					Size:   size,
				}
				
				resp, err := client.CreateImage(ctx, req)
				if err != nil {
					t.Fatalf("CreateImage with size %s failed: %v", size, err)
				}
				
				if len(resp.Data) == 0 || resp.Data[0].URL == "" {
					t.Errorf("Invalid response for size %s", size)
				}
				
				t.Logf("Image generation with size %s successful", size)
			})
		}
	})
}

// TestIntegration_ErrorScenarios tests various error scenarios with real API calls
func TestIntegration_ErrorScenarios(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("invalid_model", func(t *testing.T) {
		client := createTestClient(t)
		
		req := ChatCompletionRequest{
			Model: "invalid/nonexistent-model",
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}
		
		_, err := client.CreateChatCompletion(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid model, but got none")
		}
		
		// Check if it's a proper OpenRouter error
		if orErr, ok := err.(*OpenRouterError); ok {
			if orErr.StatusCode != 400 && orErr.StatusCode != 404 {
				t.Errorf("Expected 400 or 404 status code for invalid model, got %d", orErr.StatusCode)
			}
			t.Logf("Invalid model error handled correctly: %s", orErr.GetUserMessage())
		} else {
			t.Errorf("Expected OpenRouterError, got %T: %v", err, err)
		}
	})
	
	t.Run("invalid_api_key", func(t *testing.T) {
		config := ClientConfig{
			APIKey:   "sk-or-v1-invalid-key-12345",
			BaseURL:  DefaultBaseURL,
			SiteURL:  testSiteURL,
			SiteName: testSiteName,
			Logger:   DefaultLogger(),
		}
		
		client := NewClientWithConfig(config)
		
		req := ChatCompletionRequest{
			Model: testModel,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}
		
		_, err := client.CreateChatCompletion(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid API key, but got none")
		}
		
		// Check if it's a proper authentication error
		if orErr, ok := err.(*OpenRouterError); ok {
			if orErr.StatusCode != 401 {
				t.Errorf("Expected 401 status code for invalid API key, got %d", orErr.StatusCode)
			}
			if orErr.IsRetryable {
				t.Error("Authentication errors should not be retryable")
			}
			t.Logf("Invalid API key error handled correctly: %s", orErr.GetUserMessage())
		} else {
			t.Errorf("Expected OpenRouterError, got %T: %v", err, err)
		}
	})
	
	t.Run("empty_prompt", func(t *testing.T) {
		client := createTestClient(t)
		
		req := ChatCompletionRequest{
			Model: testModel,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: "",
				},
			},
		}
		
		_, err := client.CreateChatCompletion(ctx, req)
		if err == nil {
			t.Error("Expected validation error for empty prompt, but got none")
		}
		
		t.Logf("Empty prompt validation error: %v", err)
	})
	
	t.Run("context_timeout", func(t *testing.T) {
		client := createTestClient(t)
		
		// Create a very short timeout context
		shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		req := ChatCompletionRequest{
			Model: testModel,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}
		
		_, err := client.CreateChatCompletion(shortCtx, req)
		if err == nil {
			t.Error("Expected timeout error, but got none")
		}
		
		// Check if it's a context timeout error
		if err == context.DeadlineExceeded {
			t.Logf("Context timeout handled correctly: %v", err)
		} else {
			t.Logf("Timeout error (may be network-related): %v", err)
		}
	})
}

// TestIntegration_RateLimiting tests rate limiting behavior (if applicable)
func TestIntegration_RateLimiting(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	t.Run("rapid_requests", func(t *testing.T) {
		// Make several rapid requests to potentially trigger rate limiting
		const numRequests = 5
		
		for i := 0; i < numRequests; i++ {
			req := ChatCompletionRequest{
				Model: testModel,
				Messages: []ChatCompletionMessage{
					{
						Role:    "user",
						Content: "Quick test message " + string(rune(i+'1')),
					},
				},
			}
			
			resp, err := client.CreateChatCompletion(ctx, req)
			if err != nil {
				// Check if it's a rate limit error
				if orErr, ok := err.(*OpenRouterError); ok && orErr.StatusCode == 429 {
					t.Logf("Rate limit encountered on request %d: %s", i+1, orErr.GetUserMessage())
					
					if !orErr.IsRetryable {
						t.Error("Rate limit errors should be retryable")
					}
					
					// Don't fail the test for rate limits, just log them
					return
				} else {
					t.Fatalf("Request %d failed with non-rate-limit error: %v", i+1, err)
				}
			}
			
			if resp == nil || len(resp.Choices) == 0 {
				t.Fatalf("Request %d returned invalid response", i+1)
			}
			
			t.Logf("Request %d successful", i+1)
			
			// Small delay between requests
			time.Sleep(100 * time.Millisecond)
		}
		
		t.Logf("All %d rapid requests completed successfully", numRequests)
	})
}

// TestIntegration_ModelListing tests the model listing functionality
func TestIntegration_ModelListing(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("list_models", func(t *testing.T) {
		resp, err := client.ListModels(ctx)
		if err != nil {
			t.Fatalf("ListModels failed: %v", err)
		}
		
		if resp.Object != "list" {
			t.Errorf("Expected object 'list', got '%s'", resp.Object)
		}
		
		if len(resp.Data) == 0 {
			t.Error("No models returned")
		}
		
		// Check if our test models are available
		modelFound := false
		for _, model := range resp.Data {
			if model.ID == testModel {
				modelFound = true
				break
			}
		}
		
		if !modelFound {
			t.Logf("Warning: Test model %s not found in available models", testModel)
		}
		
		t.Logf("ListModels successful: %d models available", len(resp.Data))
	})
	
	t.Run("get_specific_model", func(t *testing.T) {
		model, err := client.GetModel(ctx, testModel)
		if err != nil {
			t.Fatalf("GetModel failed: %v", err)
		}
		
		if model.ID != testModel {
			t.Errorf("Expected model ID '%s', got '%s'", testModel, model.ID)
		}
		
		if model.Object != "model" {
			t.Errorf("Expected object 'model', got '%s'", model.Object)
		}
		
		t.Logf("GetModel successful: %s", model.ID)
	})
}

// TestIntegration_ConnectionTest tests the connection testing functionality
func TestIntegration_ConnectionTest(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("ping_api", func(t *testing.T) {
		err := client.Ping(ctx)
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
		
		t.Logf("API ping successful")
	})
	
	t.Run("ping_with_invalid_key", func(t *testing.T) {
		config := ClientConfig{
			APIKey:   "sk-or-v1-invalid-key-12345",
			BaseURL:  DefaultBaseURL,
			SiteURL:  testSiteURL,
			SiteName: testSiteName,
			Logger:   DefaultLogger(),
		}
		
		invalidClient := NewClientWithConfig(config)
		
		err := invalidClient.Ping(ctx)
		if err == nil {
			t.Error("Expected ping to fail with invalid API key")
		}
		
		t.Logf("Ping with invalid key failed as expected: %v", err)
	})
}

// TestIntegration_RetryLogic tests the retry logic with real scenarios
func TestIntegration_RetryLogic(t *testing.T) {
	client := createTestClient(t)
	
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	
	t.Run("retry_with_valid_request", func(t *testing.T) {
		retryConfig := &RetryConfig{
			MaxRetries:    2,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffFactor: 2.0,
			JitterEnabled: false,
		}
		
		var attempts int
		err := client.WithRetry(ctx, retryConfig, func() error {
			attempts++
			
			req := ChatCompletionRequest{
				Model: testModel,
				Messages: []ChatCompletionMessage{
					{
						Role:    "user",
						Content: "Test retry logic",
					},
				},
			}
			
			_, err := client.CreateChatCompletion(ctx, req)
			return err
		})
		
		if err != nil {
			t.Fatalf("Retry logic failed: %v", err)
		}
		
		if attempts != 1 {
			t.Errorf("Expected 1 attempt for successful request, got %d", attempts)
		}
		
		t.Logf("Retry logic test successful with %d attempts", attempts)
	})
}