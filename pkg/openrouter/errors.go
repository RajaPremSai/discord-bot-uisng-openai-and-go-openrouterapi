package openrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OpenRouterError represents a structured error from OpenRouter API
type OpenRouterError struct {
	StatusCode   int
	ErrorCode    string
	ErrorType    string
	Message      string
	UserMessage  string
	IsRetryable  bool
	RetryAfter   time.Duration
	OriginalErr  error
}

// Error implements the error interface
func (e *OpenRouterError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("OpenRouter API error (status: %d, code: %s): %s (original: %v)", 
			e.StatusCode, e.ErrorCode, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("OpenRouter API error (status: %d, code: %s): %s", 
		e.StatusCode, e.ErrorCode, e.Message)
}

// IsTemporary returns true if the error is temporary and should be retried
func (e *OpenRouterError) IsTemporary() bool {
	return e.IsRetryable
}

// GetUserMessage returns a user-friendly error message
func (e *OpenRouterError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return e.Message
}

// RetryConfig defines configuration for retry logic
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterEnabled   bool
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: true,
	}
}

// ParseError parses an HTTP response and returns a structured OpenRouterError
func ParseError(resp *http.Response, body []byte) *OpenRouterError {
	orErr := &OpenRouterError{
		StatusCode: resp.StatusCode,
	}

	// Try to parse JSON error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrorDetail.Message != "" {
		orErr.ErrorCode = errorResp.ErrorDetail.Code
		orErr.ErrorType = errorResp.ErrorDetail.Type
		orErr.Message = errorResp.ErrorDetail.Message
	} else {
		// Fallback to plain text or status code
		if len(body) > 0 {
			orErr.Message = string(body)
		} else {
			orErr.Message = http.StatusText(resp.StatusCode)
		}
	}

	// Set retry behavior and user messages based on status code and error type
	orErr.IsRetryable, orErr.UserMessage, orErr.RetryAfter = categorizeError(resp.StatusCode, orErr.ErrorCode, orErr.ErrorType, resp.Header)

	return orErr
}

// categorizeError determines if an error is retryable and provides user-friendly messages
func categorizeError(statusCode int, errorCode, errorType string, headers http.Header) (bool, string, time.Duration) {
	var isRetryable bool
	var userMessage string
	var retryAfter time.Duration

	switch statusCode {
	case http.StatusUnauthorized: // 401
		userMessage = "Authentication failed. Please check your OpenRouter API key."
		isRetryable = false

	case http.StatusForbidden: // 403
		if strings.Contains(strings.ToLower(errorCode), "insufficient") || 
		   strings.Contains(strings.ToLower(errorCode), "credit") ||
		   strings.Contains(strings.ToLower(errorCode), "balance") {
			userMessage = "Insufficient credits. Please add credits to your OpenRouter account."
		} else {
			userMessage = "Access forbidden. Please check your API permissions."
		}
		isRetryable = false

	case http.StatusNotFound: // 404
		if strings.Contains(strings.ToLower(errorCode), "model") {
			userMessage = "The requested AI model is not available. Please try a different model."
		} else {
			userMessage = "The requested resource was not found."
		}
		isRetryable = false

	case http.StatusTooManyRequests: // 429
		userMessage = "Rate limit exceeded. Please wait a moment before trying again."
		isRetryable = true
		
		// Parse Retry-After header if present
		if retryAfterStr := headers.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		if retryAfter == 0 {
			retryAfter = 60 * time.Second // Default retry after 1 minute
		}

	case http.StatusBadRequest: // 400
		switch strings.ToLower(errorCode) {
		case "invalid_request_error":
			userMessage = "Invalid request. Please check your input parameters."
		case "model_not_found":
			userMessage = "The specified AI model was not found. Please check the model name."
		case "context_length_exceeded":
			userMessage = "Your message is too long. Please try with a shorter message."
		default:
			userMessage = "Bad request. Please check your input and try again."
		}
		isRetryable = false

	case http.StatusInternalServerError: // 500
		userMessage = "OpenRouter service is temporarily unavailable. Please try again in a few moments."
		isRetryable = true

	case http.StatusBadGateway: // 502
		userMessage = "OpenRouter gateway error. Please try again in a few moments."
		isRetryable = true

	case http.StatusServiceUnavailable: // 503
		userMessage = "OpenRouter service is temporarily unavailable. Please try again later."
		isRetryable = true
		retryAfter = 30 * time.Second

	case http.StatusGatewayTimeout: // 504
		userMessage = "Request timed out. Please try again."
		isRetryable = true

	default:
		if statusCode >= 500 {
			userMessage = "OpenRouter service error. Please try again later."
			isRetryable = true
		} else if statusCode >= 400 {
			userMessage = "Request error. Please check your input and try again."
			isRetryable = false
		} else {
			userMessage = "An unexpected error occurred."
			isRetryable = false
		}
	}

	// Handle specific error types
	switch strings.ToLower(errorType) {
	case "insufficient_quota":
		userMessage = "Insufficient quota. Please check your OpenRouter account limits."
		isRetryable = false
	case "model_overloaded":
		userMessage = "The AI model is currently overloaded. Please try again in a few moments."
		isRetryable = true
		retryAfter = 30 * time.Second
	case "model_unavailable":
		userMessage = "The AI model is temporarily unavailable. Please try a different model or wait a few moments."
		isRetryable = true
		retryAfter = 60 * time.Second
	}

	return isRetryable, userMessage, retryAfter
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// WithRetry executes a function with exponential backoff retry logic
func WithRetry(ctx context.Context, config *RetryConfig, logger *Logger, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Check if the error is retryable
		if orErr, ok := err.(*OpenRouterError); ok {
			if !orErr.IsRetryable {
				return err // Don't retry non-retryable errors
			}

			// Log rate limit hits
			if orErr.StatusCode == http.StatusTooManyRequests && logger != nil {
				logger.LogRateLimitHit(orErr.RetryAfter)
			}

			// Use the retry-after duration if specified
			if orErr.RetryAfter > 0 {
				if logger != nil {
					logger.LogRetryAttempt(attempt+1, config.MaxRetries, orErr.RetryAfter, err)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(orErr.RetryAfter):
					continue
				}
			}
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, config)

		// Log retry attempt
		if logger != nil {
			logger.LogRetryAttempt(attempt+1, config.MaxRetries, delay, err)
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateDelay calculates the delay for the given attempt using exponential backoff
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	// Calculate exponential backoff delay
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	
	// Apply maximum delay limit
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if config.JitterEnabled {
		jitter := delay * 0.1 * (2*rand.Float64() - 1) // ±10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

// WrapNetworkError wraps network-level errors (connection, timeout, etc.)
func WrapNetworkError(err error) *OpenRouterError {
	return &OpenRouterError{
		StatusCode:  0,
		ErrorCode:   "network_error",
		ErrorType:   "network_error",
		Message:     err.Error(),
		UserMessage: "Network error occurred. Please check your internet connection and try again.",
		IsRetryable: true,
		OriginalErr: err,
	}
}

// WrapContextError wraps context-related errors (timeout, cancellation)
func WrapContextError(err error) *OpenRouterError {
	userMsg := "Request was cancelled."
	if err == context.DeadlineExceeded {
		userMsg = "Request timed out. Please try again."
	}

	return &OpenRouterError{
		StatusCode:  0,
		ErrorCode:   "context_error",
		ErrorType:   "context_error",
		Message:     err.Error(),
		UserMessage: userMsg,
		IsRetryable: err == context.DeadlineExceeded, // Retry timeouts but not cancellations
		OriginalErr: err,
	}
}

// IsRetryableError checks if an error should be retried
func IsRetryableError(err error) bool {
	if orErr, ok := err.(*OpenRouterError); ok {
		return orErr.IsRetryable
	}
	return false
}

// GetUserFriendlyMessage extracts a user-friendly message from any error
func GetUserFriendlyMessage(err error) string {
	if orErr, ok := err.(*OpenRouterError); ok {
		return orErr.GetUserMessage()
	}
	return "An unexpected error occurred. Please try again."
}