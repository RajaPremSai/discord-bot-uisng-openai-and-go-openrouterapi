package openrouter

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		headers        map[string]string
		expectedCode   string
		expectedType   string
		expectedMsg    string
		expectedUser   string
		expectedRetry  bool
		expectedAfter  time.Duration
	}{
		{
			name:          "JSON error response",
			statusCode:    400,
			body:          `{"error":{"code":"invalid_request_error","message":"Invalid model specified","type":"invalid_request_error"}}`,
			expectedCode:  "invalid_request_error",
			expectedType:  "invalid_request_error",
			expectedMsg:   "Invalid model specified",
			expectedUser:  "Invalid request. Please check your input parameters.",
			expectedRetry: false,
		},
		{
			name:          "Rate limit with retry-after",
			statusCode:    429,
			body:          `{"error":{"code":"rate_limit_exceeded","message":"Rate limit exceeded","type":"rate_limit_exceeded"}}`,
			headers:       map[string]string{"Retry-After": "60"},
			expectedCode:  "rate_limit_exceeded",
			expectedType:  "rate_limit_exceeded",
			expectedMsg:   "Rate limit exceeded",
			expectedUser:  "Rate limit exceeded. Please wait a moment before trying again.",
			expectedRetry: true,
			expectedAfter: 60 * time.Second,
		},
		{
			name:          "Unauthorized error",
			statusCode:    401,
			body:          `{"error":{"code":"invalid_api_key","message":"Invalid API key","type":"authentication_error"}}`,
			expectedCode:  "invalid_api_key",
			expectedType:  "authentication_error",
			expectedMsg:   "Invalid API key",
			expectedUser:  "Authentication failed. Please check your OpenRouter API key.",
			expectedRetry: false,
		},
		{
			name:          "Model not found",
			statusCode:    404,
			body:          `{"error":{"code":"model_not_found","message":"Model not found","type":"invalid_request_error"}}`,
			expectedCode:  "model_not_found",
			expectedType:  "invalid_request_error",
			expectedMsg:   "Model not found",
			expectedUser:  "The requested AI model is not available. Please try a different model.",
			expectedRetry: false,
		},
		{
			name:          "Insufficient credits",
			statusCode:    403,
			body:          `{"error":{"code":"insufficient_credits","message":"Insufficient credits","type":"billing_error"}}`,
			expectedCode:  "insufficient_credits",
			expectedType:  "billing_error",
			expectedMsg:   "Insufficient credits",
			expectedUser:  "Insufficient credits. Please add credits to your OpenRouter account.",
			expectedRetry: false,
		},
		{
			name:          "Server error",
			statusCode:    500,
			body:          `{"error":{"code":"internal_error","message":"Internal server error","type":"server_error"}}`,
			expectedCode:  "internal_error",
			expectedType:  "server_error",
			expectedMsg:   "Internal server error",
			expectedUser:  "OpenRouter service is temporarily unavailable. Please try again in a few moments.",
			expectedRetry: true,
		},
		{
			name:          "Plain text error",
			statusCode:    502,
			body:          "Bad Gateway",
			expectedCode:  "",
			expectedType:  "",
			expectedMsg:   "Bad Gateway",
			expectedUser:  "OpenRouter gateway error. Please try again in a few moments.",
			expectedRetry: true,
		},
		{
			name:          "Empty body",
			statusCode:    503,
			body:          "",
			expectedCode:  "",
			expectedType:  "",
			expectedMsg:   "Service Unavailable",
			expectedUser:  "OpenRouter service is temporarily unavailable. Please try again later.",
			expectedRetry: true,
			expectedAfter: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock response
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}

			// Add headers if specified
			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			// Parse the error
			orErr := ParseError(resp, []byte(tt.body))

			// Verify results
			if orErr.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, orErr.StatusCode)
			}
			if orErr.ErrorCode != tt.expectedCode {
				t.Errorf("Expected error code %q, got %q", tt.expectedCode, orErr.ErrorCode)
			}
			if orErr.ErrorType != tt.expectedType {
				t.Errorf("Expected error type %q, got %q", tt.expectedType, orErr.ErrorType)
			}
			if orErr.Message != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, orErr.Message)
			}
			if orErr.UserMessage != tt.expectedUser {
				t.Errorf("Expected user message %q, got %q", tt.expectedUser, orErr.UserMessage)
			}
			if orErr.IsRetryable != tt.expectedRetry {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetry, orErr.IsRetryable)
			}
			if orErr.RetryAfter != tt.expectedAfter {
				t.Errorf("Expected retry after %v, got %v", tt.expectedAfter, orErr.RetryAfter)
			}
		})
	}
}

func TestOpenRouterError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *OpenRouterError
		expected string
	}{
		{
			name: "Basic error",
			err: &OpenRouterError{
				StatusCode: 400,
				ErrorCode:  "invalid_request",
				Message:    "Invalid request",
			},
			expected: "OpenRouter API error (status: 400, code: invalid_request): Invalid request",
		},
		{
			name: "Error with original error",
			err: &OpenRouterError{
				StatusCode:  500,
				ErrorCode:   "internal_error",
				Message:     "Internal error",
				OriginalErr: errors.New("connection failed"),
			},
			expected: "OpenRouter API error (status: 500, code: internal_error): Internal error (original: connection failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestOpenRouterError_GetUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *OpenRouterError
		expected string
	}{
		{
			name: "With user message",
			err: &OpenRouterError{
				Message:     "Technical error message",
				UserMessage: "User-friendly message",
			},
			expected: "User-friendly message",
		},
		{
			name: "Without user message",
			err: &OpenRouterError{
				Message: "Technical error message",
			},
			expected: "Technical error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.GetUserMessage(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return &OpenRouterError{IsRetryable: true}
		}
		return nil
	}

	config := &RetryConfig{
		MaxRetries:    5,
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	ctx := context.Background()
	err := WithRetry(ctx, config, fn)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	callCount := 0
	expectedErr := &OpenRouterError{
		IsRetryable: false,
		Message:     "Non-retryable error",
	}

	fn := func() error {
		callCount++
		return expectedErr
	}

	config := &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	ctx := context.Background()
	err := WithRetry(ctx, config, fn)

	if err != expectedErr {
		t.Errorf("Expected specific error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return &OpenRouterError{IsRetryable: true, Message: "Always fails"}
	}

	config := &RetryConfig{
		MaxRetries:    2,
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	ctx := context.Background()
	err := WithRetry(ctx, config, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if callCount != 3 { // Initial call + 2 retries
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return &OpenRouterError{IsRetryable: true}
	}

	config := &RetryConfig{
		MaxRetries:    5,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := WithRetry(ctx, config, fn)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_RetryAfter(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 2 {
			return &OpenRouterError{
				IsRetryable: true,
				RetryAfter:  10 * time.Millisecond,
			}
		}
		return nil
	}

	config := &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     100 * time.Millisecond, // This should be ignored due to RetryAfter
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
	}

	ctx := context.Background()
	start := time.Now()
	err := WithRetry(ctx, config, fn)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
	// Should have waited at least 10ms for RetryAfter
	if duration < 10*time.Millisecond {
		t.Errorf("Expected to wait at least 10ms, waited %v", duration)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: false, // Disable jitter for predictable testing
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxDelay
		{5, 1 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := calculateDelay(tt.attempt, config)
			if delay != tt.expected {
				t.Errorf("Attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
			}
		})
	}
}

func TestWrapNetworkError(t *testing.T) {
	originalErr := errors.New("connection refused")
	orErr := WrapNetworkError(originalErr)

	if orErr.StatusCode != 0 {
		t.Errorf("Expected status code 0, got %d", orErr.StatusCode)
	}
	if orErr.ErrorCode != "network_error" {
		t.Errorf("Expected error code 'network_error', got %q", orErr.ErrorCode)
	}
	if orErr.ErrorType != "network_error" {
		t.Errorf("Expected error type 'network_error', got %q", orErr.ErrorType)
	}
	if !orErr.IsRetryable {
		t.Error("Expected network error to be retryable")
	}
	if orErr.OriginalErr != originalErr {
		t.Errorf("Expected original error to be preserved")
	}
	if !strings.Contains(orErr.UserMessage, "Network error") {
		t.Errorf("Expected user message to mention network error, got %q", orErr.UserMessage)
	}
}

func TestWrapContextError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
		retryable   bool
	}{
		{
			name:        "Deadline exceeded",
			err:         context.DeadlineExceeded,
			expectedMsg: "Request timed out. Please try again.",
			retryable:   true,
		},
		{
			name:        "Context canceled",
			err:         context.Canceled,
			expectedMsg: "Request was cancelled.",
			retryable:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orErr := WrapContextError(tt.err)

			if orErr.ErrorCode != "context_error" {
				t.Errorf("Expected error code 'context_error', got %q", orErr.ErrorCode)
			}
			if orErr.UserMessage != tt.expectedMsg {
				t.Errorf("Expected user message %q, got %q", tt.expectedMsg, orErr.UserMessage)
			}
			if orErr.IsRetryable != tt.retryable {
				t.Errorf("Expected retryable %v, got %v", tt.retryable, orErr.IsRetryable)
			}
			if orErr.OriginalErr != tt.err {
				t.Error("Expected original error to be preserved")
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Retryable OpenRouter error",
			err:      &OpenRouterError{IsRetryable: true},
			expected: true,
		},
		{
			name:     "Non-retryable OpenRouter error",
			err:      &OpenRouterError{IsRetryable: false},
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "OpenRouter error with user message",
			err: &OpenRouterError{
				Message:     "Technical message",
				UserMessage: "User-friendly message",
			},
			expected: "User-friendly message",
		},
		{
			name: "OpenRouter error without user message",
			err: &OpenRouterError{
				Message: "Technical message",
			},
			expected: "Technical message",
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: "An unexpected error occurred. Please try again.",
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: "An unexpected error occurred. Please try again.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserFriendlyMessage(tt.err); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay 1s, got %v", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay 30s, got %v", config.MaxDelay)
	}
	if config.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor 2.0, got %f", config.BackoffFactor)
	}
	if !config.JitterEnabled {
		t.Error("Expected JitterEnabled to be true")
	}
}

// Benchmark tests for performance validation
func BenchmarkParseError(b *testing.B) {
	resp := &http.Response{
		StatusCode: 400,
		Header:     make(http.Header),
	}
	body := []byte(`{"error":{"code":"invalid_request_error","message":"Invalid model specified","type":"invalid_request_error"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseError(resp, body)
	}
}

func BenchmarkCalculateDelay(b *testing.B) {
	config := DefaultRetryConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateDelay(i%10, config)
	}
}