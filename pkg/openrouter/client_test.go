package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, client.apiKey)
	}

	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected base URL %s, got %s", DefaultBaseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	config := ClientConfig{
		APIKey:   "test-api-key",
		BaseURL:  "https://custom.api.com/v1",
		SiteURL:  "https://example.com",
		SiteName: "Test Bot",
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	client := NewClientWithConfig(config)

	if client.apiKey != config.APIKey {
		t.Errorf("Expected API key %s, got %s", config.APIKey, client.apiKey)
	}

	if client.baseURL != config.BaseURL {
		t.Errorf("Expected base URL %s, got %s", config.BaseURL, client.baseURL)
	}

	if client.siteURL != config.SiteURL {
		t.Errorf("Expected site URL %s, got %s", config.SiteURL, client.siteURL)
	}

	if client.siteName != config.SiteName {
		t.Errorf("Expected site name %s, got %s", config.SiteName, client.siteName)
	}

	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("Expected timeout %v, got %v", 60*time.Second, client.httpClient.Timeout)
	}
}

func TestNewClientWithConfigDefaults(t *testing.T) {
	config := ClientConfig{
		APIKey: "test-api-key",
	}

	client := NewClientWithConfig(config)

	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected default base URL %s, got %s", DefaultBaseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized with default")
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected default timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

func TestSetSiteInfo(t *testing.T) {
	client := NewClient("test-api-key")
	siteURL := "https://example.com"
	siteName := "Test Bot"

	client.SetSiteInfo(siteURL, siteName)

	if client.siteURL != siteURL {
		t.Errorf("Expected site URL %s, got %s", siteURL, client.siteURL)
	}

	if client.siteName != siteName {
		t.Errorf("Expected site name %s, got %s", siteName, client.siteName)
	}
}

func TestBuildRequest(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		method   string
		endpoint string
		body     interface{}
		wantURL  string
		wantHeaders map[string]string
	}{
		{
			name: "basic request with required headers",
			client: &Client{
				apiKey:  "test-api-key",
				baseURL: "https://openrouter.ai/api/v1",
			},
			method:   "POST",
			endpoint: "/chat/completions",
			body:     map[string]string{"test": "data"},
			wantURL:  "https://openrouter.ai/api/v1/chat/completions",
			wantHeaders: map[string]string{
				"Authorization": "Bearer test-api-key",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "request with optional headers",
			client: &Client{
				apiKey:   "test-api-key",
				baseURL:  "https://openrouter.ai/api/v1",
				siteURL:  "https://example.com",
				siteName: "Test Bot",
			},
			method:   "GET",
			endpoint: "/models",
			body:     nil,
			wantURL:  "https://openrouter.ai/api/v1/models",
			wantHeaders: map[string]string{
				"Authorization": "Bearer test-api-key",
				"Content-Type":  "application/json",
				"HTTP-Referer":  "https://example.com",
				"X-Title":       "Test Bot",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req, err := tt.client.buildRequest(ctx, tt.method, tt.endpoint, tt.body)
			if err != nil {
				t.Fatalf("buildRequest() error = %v", err)
			}

			if req.URL.String() != tt.wantURL {
				t.Errorf("Expected URL %s, got %s", tt.wantURL, req.URL.String())
			}

			if req.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, req.Method)
			}

			for key, expectedValue := range tt.wantHeaders {
				actualValue := req.Header.Get(key)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s: %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestBuildRequestWithInvalidJSON(t *testing.T) {
	client := NewClient("test-api-key")
	ctx := context.Background()

	// Use a channel which cannot be marshaled to JSON
	invalidBody := make(chan int)

	_, err := client.buildRequest(ctx, "POST", "/test", invalidBody)
	if err == nil {
		t.Error("Expected error for invalid JSON body")
	}

	if !strings.Contains(err.Error(), "failed to marshal request body") {
		t.Errorf("Expected marshal error, got: %v", err)
	}
}

func TestDoRequestSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", auth)
		}

		response := map[string]interface{}{
			"id":     "test-id",
			"object": "chat.completion",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req, err := client.buildRequest(ctx, "GET", "/test", nil)
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}

	var result map[string]interface{}
	err = client.doRequest(req, &result)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}

	if result["id"] != "test-id" {
		t.Errorf("Expected id 'test-id', got %v", result["id"])
	}
}

func TestDoRequestError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "invalid_request",
				Message: "Invalid request parameters",
				Type:    "invalid_request_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req, err := client.buildRequest(ctx, "POST", "/test", nil)
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}

	err = client.doRequest(req, nil)
	if err == nil {
		t.Error("Expected error for 400 status code")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "invalid_request" {
		t.Errorf("Expected error code 'invalid_request', got '%s'", errorResp.ErrorDetail.Code)
	}
}

func TestPing(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("Expected path '/models', got '%s'", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", r.Method)
		}

		response := ModelsResponse{
			Object: "list",
			Data:   []Model{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	err := client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestCreateChatCompletionHeaders(t *testing.T) {
	// Create a test server to verify headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all required headers are present
		expectedHeaders := map[string]string{
			"Authorization": "Bearer test-api-key",
			"Content-Type":  "application/json",
			"HTTP-Referer":  "https://example.com",
			"X-Title":       "Test Bot",
		}

		for key, expectedValue := range expectedHeaders {
			actualValue := r.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s: %s, got %s", key, expectedValue, actualValue)
			}
		}

		response := ChatCompletionResponse{
			ID:     "test-id",
			Object: "chat.completion",
			Model:  "openai/gpt-3.5-turbo",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:   "test-api-key",
		BaseURL:  server.URL,
		SiteURL:  "https://example.com",
		SiteName: "Test Bot",
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Errorf("CreateChatCompletion() error = %v", err)
	}
}

func TestCreateChatCompletionSuccess(t *testing.T) {
	// Create a test server that returns a successful chat completion response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path '/chat/completions', got %s", r.URL.Path)
		}

		// Verify request body
		var reqBody ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Model != "openai/gpt-4" {
			t.Errorf("Expected model 'openai/gpt-4', got %s", reqBody.Model)
		}

		if len(reqBody.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(reqBody.Messages))
		}

		// Return successful response
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test123",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "openai/gpt-4",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     9,
				CompletionTokens: 12,
				TotalTokens:      21,
				PromptCost:       0.00027,
				CompletionCost:   0.00036,
				TotalCost:        0.00063,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(100),
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error = %v", err)
	}

	// Verify response
	if resp.ID != "chatcmpl-test123" {
		t.Errorf("Expected ID 'chatcmpl-test123', got %s", resp.ID)
	}

	if resp.Model != "openai/gpt-4" {
		t.Errorf("Expected model 'openai/gpt-4', got %s", resp.Model)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected content 'Hello! How can I help you today?', got %s", resp.Choices[0].Message.Content)
	}

	if resp.Usage.TotalTokens != 21 {
		t.Errorf("Expected total tokens 21, got %d", resp.Usage.TotalTokens)
	}
}

func TestCreateChatCompletionValidationError(t *testing.T) {
	client := NewClient("test-api-key")
	ctx := context.Background()

	tests := []struct {
		name    string
		request ChatCompletionRequest
		wantErr string
	}{
		{
			name: "missing model",
			request: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: "invalid request: model is required",
		},
		{
			name: "no messages",
			request: ChatCompletionRequest{
				Model:    "openai/gpt-4",
				Messages: []ChatCompletionMessage{},
			},
			wantErr: "invalid request: at least one message is required",
		},
		{
			name: "message missing role",
			request: ChatCompletionRequest{
				Model: "openai/gpt-4",
				Messages: []ChatCompletionMessage{
					{Content: "Hello"},
				},
			},
			wantErr: "invalid request: message 0: role is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateChatCompletion(ctx, tt.request)
			if err == nil {
				t.Error("Expected validation error but got none")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCreateChatCompletionAPIError(t *testing.T) {
	// Create a test server that returns an API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "invalid_request_error",
				Message: "Invalid model specified",
				Type:    "invalid_request_error",
				Param:   "model",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "invalid-model",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := client.CreateChatCompletion(ctx, req)
	if err == nil {
		t.Error("Expected API error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "invalid_request_error" {
		t.Errorf("Expected error code 'invalid_request_error', got %s", errorResp.ErrorDetail.Code)
	}

	if errorResp.ErrorDetail.Message != "Invalid model specified" {
		t.Errorf("Expected error message 'Invalid model specified', got %s", errorResp.ErrorDetail.Message)
	}
}

func TestCreateChatCompletionRateLimitError(t *testing.T) {
	// Create a test server that returns a rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "rate_limit_exceeded",
				Message: "Rate limit exceeded. Please try again later.",
				Type:    "rate_limit_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := client.CreateChatCompletion(ctx, req)
	if err == nil {
		t.Error("Expected rate limit error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "rate_limit_exceeded" {
		t.Errorf("Expected error code 'rate_limit_exceeded', got %s", errorResp.ErrorDetail.Code)
	}
}

func TestCreateChatCompletionAuthenticationError(t *testing.T) {
	// Create a test server that returns an authentication error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "invalid_api_key",
				Message: "Invalid API key provided",
				Type:    "authentication_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "invalid-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := client.CreateChatCompletion(ctx, req)
	if err == nil {
		t.Error("Expected authentication error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "invalid_api_key" {
		t.Errorf("Expected error code 'invalid_api_key', got %s", errorResp.ErrorDetail.Code)
	}
}

func TestCreateChatCompletionWithAllParameters(t *testing.T) {
	// Create a test server that verifies all parameters are sent correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify all parameters
		if reqBody.Model != "openai/gpt-3.5-turbo" {
			t.Errorf("Expected model 'openai/gpt-3.5-turbo', got %s", reqBody.Model)
		}

		if reqBody.Temperature == nil || *reqBody.Temperature != 0.8 {
			t.Errorf("Expected temperature 0.8, got %v", reqBody.Temperature)
		}

		if reqBody.MaxTokens == nil || *reqBody.MaxTokens != 150 {
			t.Errorf("Expected max_tokens 150, got %v", reqBody.MaxTokens)
		}

		if reqBody.TopP == nil || *reqBody.TopP != 0.9 {
			t.Errorf("Expected top_p 0.9, got %v", reqBody.TopP)
		}

		if reqBody.FrequencyPenalty == nil || *reqBody.FrequencyPenalty != 0.1 {
			t.Errorf("Expected frequency_penalty 0.1, got %v", reqBody.FrequencyPenalty)
		}

		if reqBody.PresencePenalty == nil || *reqBody.PresencePenalty != 0.2 {
			t.Errorf("Expected presence_penalty 0.2, got %v", reqBody.PresencePenalty)
		}

		if len(reqBody.Stop) != 2 || reqBody.Stop[0] != "END" || reqBody.Stop[1] != "STOP" {
			t.Errorf("Expected stop tokens ['END', 'STOP'], got %v", reqBody.Stop)
		}

		if reqBody.User != "test-user-123" {
			t.Errorf("Expected user 'test-user-123', got %s", reqBody.User)
		}

		// Return successful response
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test456",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "openai/gpt-3.5-turbo",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "This is a test response with all parameters.",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     25,
				CompletionTokens: 10,
				TotalTokens:      35,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "Hello, how are you?",
				Name:    "user1",
			},
		},
		Temperature:      floatPtr(0.8),
		MaxTokens:        intPtr(150),
		TopP:             floatPtr(0.9),
		FrequencyPenalty: floatPtr(0.1),
		PresencePenalty:  floatPtr(0.2),
		Stop:             []string{"END", "STOP"},
		User:             "test-user-123",
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error = %v", err)
	}

	if resp.ID != "chatcmpl-test456" {
		t.Errorf("Expected ID 'chatcmpl-test456', got %s", resp.ID)
	}

	if resp.Choices[0].Message.Content != "This is a test response with all parameters." {
		t.Errorf("Expected specific content, got %s", resp.Choices[0].Message.Content)
	}
}

func TestCreateChatCompletionMultipleChoices(t *testing.T) {
	// Create a test server that returns multiple choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatCompletionResponse{
			ID:      "chatcmpl-multi123",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "openai/gpt-4",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "First response option.",
					},
					FinishReason: "stop",
				},
				{
					Index: 1,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "Second response option.",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 12,
				TotalTokens:      22,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Give me two options.",
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error = %v", err)
	}

	if len(resp.Choices) != 2 {
		t.Errorf("Expected 2 choices, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "First response option." {
		t.Errorf("Expected first choice content 'First response option.', got %s", resp.Choices[0].Message.Content)
	}

	if resp.Choices[1].Message.Content != "Second response option." {
		t.Errorf("Expected second choice content 'Second response option.', got %s", resp.Choices[1].Message.Content)
	}
}

func TestCreateImageHeaders(t *testing.T) {
	// Create a test server to verify headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all required headers are present
		expectedHeaders := map[string]string{
			"Authorization": "Bearer test-api-key",
			"Content-Type":  "application/json",
			"HTTP-Referer":  "https://example.com",
			"X-Title":       "Test Bot",
		}

		for key, expectedValue := range expectedHeaders {
			actualValue := r.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s: %s, got %s", key, expectedValue, actualValue)
			}
		}

		response := ImageResponse{
			Created: 1234567890,
			Data: []ImageData{
				{
					URL: "https://example.com/image.png",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:   "test-api-key",
		BaseURL:  server.URL,
		SiteURL:  "https://example.com",
		SiteName: "Test Bot",
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A beautiful sunset",
		Model:  "openai/dall-e-2",
		N:      1,
	}

	_, err := client.CreateImage(ctx, req)
	if err != nil {
		t.Errorf("CreateImage() error = %v", err)
	}
}

func TestCreateImageSuccess(t *testing.T) {
	// Create a test server that returns a successful image generation response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/images/generations" {
			t.Errorf("Expected path '/images/generations', got %s", r.URL.Path)
		}

		// Verify request body
		var reqBody ImageRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Model != "openai/dall-e-3" {
			t.Errorf("Expected model 'openai/dall-e-3', got %s", reqBody.Model)
		}

		if reqBody.Prompt != "A futuristic cityscape at night" {
			t.Errorf("Expected prompt 'A futuristic cityscape at night', got %s", reqBody.Prompt)
		}

		if reqBody.N != 1 {
			t.Errorf("Expected n=1, got %d", reqBody.N)
		}

		// Return successful response
		response := ImageResponse{
			Created: 1677652288,
			Data: []ImageData{
				{
					URL:           "https://example.com/generated-image.png",
					RevisedPrompt: "A detailed futuristic cityscape at night with neon lights and flying cars",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A futuristic cityscape at night",
		Model:  "openai/dall-e-3",
		N:      1,
		Size:   "1024x1024",
	}

	resp, err := client.CreateImage(ctx, req)
	if err != nil {
		t.Fatalf("CreateImage() error = %v", err)
	}

	// Verify response
	if resp.Created != 1677652288 {
		t.Errorf("Expected created timestamp 1677652288, got %d", resp.Created)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 image, got %d", len(resp.Data))
	}

	if resp.Data[0].URL != "https://example.com/generated-image.png" {
		t.Errorf("Expected URL 'https://example.com/generated-image.png', got %s", resp.Data[0].URL)
	}

	if resp.Data[0].RevisedPrompt != "A detailed futuristic cityscape at night with neon lights and flying cars" {
		t.Errorf("Expected revised prompt, got %s", resp.Data[0].RevisedPrompt)
	}
}

func TestCreateImageValidationError(t *testing.T) {
	client := NewClient("test-api-key")
	ctx := context.Background()

	tests := []struct {
		name    string
		request ImageRequest
		wantErr string
	}{
		{
			name: "missing prompt",
			request: ImageRequest{
				Model: "openai/dall-e-2",
				N:     1,
			},
			wantErr: "invalid request: prompt is required",
		},
		{
			name: "missing model",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				N:      1,
			},
			wantErr: "invalid request: model is required",
		},
		{
			name: "negative n",
			request: ImageRequest{
				Prompt: "A beautiful sunset",
				Model:  "openai/dall-e-2",
				N:      -1,
			},
			wantErr: "invalid request: n must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateImage(ctx, tt.request)
			if err == nil {
				t.Error("Expected validation error but got none")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCreateImageAPIError(t *testing.T) {
	// Create a test server that returns an API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "invalid_request_error",
				Message: "Invalid image model specified",
				Type:    "invalid_request_error",
				Param:   "model",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A beautiful sunset",
		Model:  "invalid-image-model",
		N:      1,
	}

	_, err := client.CreateImage(ctx, req)
	if err == nil {
		t.Error("Expected API error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "invalid_request_error" {
		t.Errorf("Expected error code 'invalid_request_error', got %s", errorResp.ErrorDetail.Code)
	}

	if errorResp.ErrorDetail.Message != "Invalid image model specified" {
		t.Errorf("Expected error message 'Invalid image model specified', got %s", errorResp.ErrorDetail.Message)
	}
}

func TestCreateImageRateLimitError(t *testing.T) {
	// Create a test server that returns a rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "rate_limit_exceeded",
				Message: "Rate limit exceeded for image generation. Please try again later.",
				Type:    "rate_limit_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A beautiful sunset",
		Model:  "openai/dall-e-3",
		N:      1,
	}

	_, err := client.CreateImage(ctx, req)
	if err == nil {
		t.Error("Expected rate limit error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "rate_limit_exceeded" {
		t.Errorf("Expected error code 'rate_limit_exceeded', got %s", errorResp.ErrorDetail.Code)
	}
}

func TestCreateImageAuthenticationError(t *testing.T) {
	// Create a test server that returns an authentication error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "invalid_api_key",
				Message: "Invalid API key provided for image generation",
				Type:    "authentication_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "invalid-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A beautiful sunset",
		Model:  "openai/dall-e-2",
		N:      1,
	}

	_, err := client.CreateImage(ctx, req)
	if err == nil {
		t.Error("Expected authentication error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "invalid_api_key" {
		t.Errorf("Expected error code 'invalid_api_key', got %s", errorResp.ErrorDetail.Code)
	}
}

func TestCreateImageWithAllParameters(t *testing.T) {
	// Create a test server that verifies all parameters are sent correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody ImageRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify all parameters
		if reqBody.Prompt != "A majestic mountain landscape" {
			t.Errorf("Expected prompt 'A majestic mountain landscape', got %s", reqBody.Prompt)
		}

		if reqBody.Model != "openai/dall-e-3" {
			t.Errorf("Expected model 'openai/dall-e-3', got %s", reqBody.Model)
		}

		if reqBody.N != 2 {
			t.Errorf("Expected n=2, got %d", reqBody.N)
		}

		if reqBody.Size != "1024x1024" {
			t.Errorf("Expected size '1024x1024', got %s", reqBody.Size)
		}

		if reqBody.ResponseFormat != "url" {
			t.Errorf("Expected response_format 'url', got %s", reqBody.ResponseFormat)
		}

		if reqBody.User != "test-user-456" {
			t.Errorf("Expected user 'test-user-456', got %s", reqBody.User)
		}

		if reqBody.Quality != "hd" {
			t.Errorf("Expected quality 'hd', got %s", reqBody.Quality)
		}

		if reqBody.Style != "vivid" {
			t.Errorf("Expected style 'vivid', got %s", reqBody.Style)
		}

		// Return successful response with multiple images
		response := ImageResponse{
			Created: 1677652288,
			Data: []ImageData{
				{
					URL:           "https://example.com/mountain1.png",
					RevisedPrompt: "A majestic mountain landscape with snow-capped peaks and a clear blue sky",
				},
				{
					URL:           "https://example.com/mountain2.png",
					RevisedPrompt: "A majestic mountain landscape with dramatic lighting and alpine vegetation",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt:         "A majestic mountain landscape",
		Model:          "openai/dall-e-3",
		N:              2,
		Size:           "1024x1024",
		ResponseFormat: "url",
		User:           "test-user-456",
		Quality:        "hd",
		Style:          "vivid",
	}

	resp, err := client.CreateImage(ctx, req)
	if err != nil {
		t.Fatalf("CreateImage() error = %v", err)
	}

	if resp.Created != 1677652288 {
		t.Errorf("Expected created timestamp 1677652288, got %d", resp.Created)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 images, got %d", len(resp.Data))
	}

	if resp.Data[0].URL != "https://example.com/mountain1.png" {
		t.Errorf("Expected first image URL 'https://example.com/mountain1.png', got %s", resp.Data[0].URL)
	}

	if resp.Data[1].URL != "https://example.com/mountain2.png" {
		t.Errorf("Expected second image URL 'https://example.com/mountain2.png', got %s", resp.Data[1].URL)
	}
}

func TestCreateImageWithBase64Response(t *testing.T) {
	// Create a test server that returns base64 encoded images
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody ImageRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.ResponseFormat != "b64_json" {
			t.Errorf("Expected response_format 'b64_json', got %s", reqBody.ResponseFormat)
		}

		// Return successful response with base64 data
		response := ImageResponse{
			Created: 1677652288,
			Data: []ImageData{
				{
					B64JSON:       "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
					RevisedPrompt: "A simple test image",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt:         "A simple test image",
		Model:          "openai/dall-e-2",
		N:              1,
		ResponseFormat: "b64_json",
	}

	resp, err := client.CreateImage(ctx, req)
	if err != nil {
		t.Fatalf("CreateImage() error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 image, got %d", len(resp.Data))
	}

	expectedB64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	if resp.Data[0].B64JSON != expectedB64 {
		t.Errorf("Expected base64 data %s, got %s", expectedB64, resp.Data[0].B64JSON)
	}

	if resp.Data[0].URL != "" {
		t.Errorf("Expected empty URL for base64 response, got %s", resp.Data[0].URL)
	}
}

func TestCreateImageModelUnavailableError(t *testing.T) {
	// Create a test server that returns a model unavailable error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		errorResp := ErrorResponse{
			ErrorDetail: ErrorDetail{
				Code:    "model_unavailable",
				Message: "The requested image model is currently unavailable. Please try again later.",
				Type:    "service_unavailable_error",
				Param:   "model",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ImageRequest{
		Prompt: "A beautiful sunset",
		Model:  "openai/dall-e-3",
		N:      1,
	}

	_, err := client.CreateImage(ctx, req)
	if err == nil {
		t.Error("Expected model unavailable error but got none")
	}

	errorResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.ErrorDetail.Code != "model_unavailable" {
		t.Errorf("Expected error code 'model_unavailable', got %s", errorResp.ErrorDetail.Code)
	}

	if errorResp.ErrorDetail.Type != "service_unavailable_error" {
		t.Errorf("Expected error type 'service_unavailable_error', got %s", errorResp.ErrorDetail.Type)
	}
}

