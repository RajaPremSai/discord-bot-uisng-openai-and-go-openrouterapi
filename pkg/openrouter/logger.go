package openrouter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles logging for OpenRouter API interactions
type Logger struct {
	level           LogLevel
	enableMetrics   bool
	enableRequestLog bool
	enableResponseLog bool
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level             LogLevel
	EnableMetrics     bool
	EnableRequestLog  bool
	EnableResponseLog bool
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config LoggerConfig) *Logger {
	return &Logger{
		level:           config.Level,
		enableMetrics:   config.EnableMetrics,
		enableRequestLog: config.EnableRequestLog,
		enableResponseLog: config.EnableResponseLog,
	}
}

// DefaultLogger returns a logger with default configuration
func DefaultLogger() *Logger {
	return &Logger{
		level:           LogLevelInfo,
		enableMetrics:   true,
		enableRequestLog: true,
		enableResponseLog: true,
	}
}

// APICallMetrics holds performance metrics for an API call
type APICallMetrics struct {
	Endpoint        string        `json:"endpoint"`
	Method          string        `json:"method"`
	Model           string        `json:"model,omitempty"`
	Duration        time.Duration `json:"duration"`
	StatusCode      int           `json:"status_code"`
	Success         bool          `json:"success"`
	RequestSize     int64         `json:"request_size,omitempty"`
	ResponseSize    int64         `json:"response_size,omitempty"`
	PromptTokens    int           `json:"prompt_tokens,omitempty"`
	CompletionTokens int          `json:"completion_tokens,omitempty"`
	TotalTokens     int           `json:"total_tokens,omitempty"`
	ErrorCode       string        `json:"error_code,omitempty"`
	ErrorType       string        `json:"error_type,omitempty"`
	Timestamp       time.Time     `json:"timestamp"`
}

// RequestLogData holds data for request logging
type RequestLogData struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      interface{}       `json:"body,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// ResponseLogData holds data for response logging
type ResponseLogData struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         interface{}       `json:"body,omitempty"`
	Duration     time.Duration     `json:"duration"`
	Success      bool              `json:"success"`
	Timestamp    time.Time         `json:"timestamp"`
}

// shouldLog checks if a message should be logged based on the current log level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// logf logs a formatted message with the given level
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}
	
	prefix := fmt.Sprintf("[OpenRouter][%s]", level.String())
	message := fmt.Sprintf(format, args...)
	log.Printf("%s %s", prefix, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(LogLevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logf(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(LogLevelError, format, args...)
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(req *http.Request, body interface{}) {
	if !l.enableRequestLog || !l.shouldLog(LogLevelDebug) {
		return
	}

	// Sanitize headers (remove sensitive information)
	headers := make(map[string]string)
	for key, values := range req.Header {
		if strings.ToLower(key) == "authorization" {
			headers[key] = "Bearer [REDACTED]"
		} else {
			headers[key] = strings.Join(values, ", ")
		}
	}

	requestData := RequestLogData{
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   headers,
		Body:      body,
		Timestamp: time.Now(),
	}

	if jsonData, err := json.Marshal(requestData); err == nil {
		l.Debug("API Request: %s", string(jsonData))
	} else {
		l.Debug("API Request: %s %s (failed to serialize request data: %v)", req.Method, req.URL.String(), err)
	}
}

// LogResponse logs an HTTP response
func (l *Logger) LogResponse(statusCode int, headers http.Header, body interface{}, duration time.Duration) {
	if !l.enableResponseLog || !l.shouldLog(LogLevelDebug) {
		return
	}

	// Convert headers to map
	headerMap := make(map[string]string)
	for key, values := range headers {
		headerMap[key] = strings.Join(values, ", ")
	}

	responseData := ResponseLogData{
		StatusCode: statusCode,
		Headers:    headerMap,
		Body:       body,
		Duration:   duration,
		Success:    statusCode >= 200 && statusCode < 300,
		Timestamp:  time.Now(),
	}

	if jsonData, err := json.Marshal(responseData); err == nil {
		l.Debug("API Response: %s", string(jsonData))
	} else {
		l.Debug("API Response: Status %d, Duration %v (failed to serialize response data: %v)", statusCode, duration, err)
	}
}

// LogMetrics logs performance metrics for an API call
func (l *Logger) LogMetrics(metrics APICallMetrics) {
	if !l.enableMetrics || !l.shouldLog(LogLevelInfo) {
		return
	}

	if jsonData, err := json.Marshal(metrics); err == nil {
		l.Info("API Metrics: %s", string(jsonData))
	} else {
		l.Info("API Metrics: %s %s - Duration: %v, Status: %d, Success: %t (failed to serialize metrics: %v)", 
			metrics.Method, metrics.Endpoint, metrics.Duration, metrics.StatusCode, metrics.Success, err)
	}
}

// LogError logs an OpenRouter API error with detailed information
func (l *Logger) LogError(err error, context string) {
	if !l.shouldLog(LogLevelError) {
		return
	}

	if orErr, ok := err.(*OpenRouterError); ok {
		l.Error("%s - OpenRouter Error: Status=%d, Code=%s, Type=%s, Message=%s, Retryable=%t", 
			context, orErr.StatusCode, orErr.ErrorCode, orErr.ErrorType, orErr.Message, orErr.IsRetryable)
		
		if orErr.OriginalErr != nil {
			l.Error("%s - Original Error: %v", context, orErr.OriginalErr)
		}
	} else {
		l.Error("%s - Error: %v", context, err)
	}
}

// LogChatCompletion logs specific information about chat completion requests
func (l *Logger) LogChatCompletion(req ChatCompletionRequest, resp *ChatCompletionResponse, duration time.Duration, err error) {
	if err != nil {
		l.LogError(err, "Chat Completion")
		return
	}

	if !l.shouldLog(LogLevelInfo) {
		return
	}

	metrics := APICallMetrics{
		Endpoint:     "/chat/completions",
		Method:       "POST",
		Model:        req.Model,
		Duration:     duration,
		StatusCode:   200,
		Success:      true,
		Timestamp:    time.Now(),
	}

	if resp != nil && resp.Usage != (Usage{}) {
		metrics.PromptTokens = resp.Usage.PromptTokens
		metrics.CompletionTokens = resp.Usage.CompletionTokens
		metrics.TotalTokens = resp.Usage.TotalTokens
	}

	l.LogMetrics(metrics)
	
	// Log additional chat-specific information
	l.Info("Chat Completion: Model=%s, Messages=%d, Temperature=%.2f, MaxTokens=%d, Duration=%v", 
		req.Model, len(req.Messages), getTemperature(req.Temperature), getMaxTokens(req.MaxTokens), duration)
}

// LogImageGeneration logs specific information about image generation requests
func (l *Logger) LogImageGeneration(req ImageRequest, resp *ImageResponse, duration time.Duration, err error) {
	if err != nil {
		l.LogError(err, "Image Generation")
		return
	}

	if !l.shouldLog(LogLevelInfo) {
		return
	}

	metrics := APICallMetrics{
		Endpoint:   "/images/generations",
		Method:     "POST",
		Model:      req.Model,
		Duration:   duration,
		StatusCode: 200,
		Success:    true,
		Timestamp:  time.Now(),
	}

	l.LogMetrics(metrics)
	
	// Log additional image-specific information
	imagesGenerated := 0
	if resp != nil {
		imagesGenerated = len(resp.Data)
	}
	
	l.Info("Image Generation: Model=%s, Prompt=%s, Size=%s, Count=%d, Generated=%d, Duration=%v", 
		req.Model, truncateString(req.Prompt, 100), req.Size, req.N, imagesGenerated, duration)
}

// LogRetryAttempt logs information about retry attempts
func (l *Logger) LogRetryAttempt(attempt int, maxRetries int, delay time.Duration, err error) {
	if !l.shouldLog(LogLevelWarn) {
		return
	}

	l.Warn("Retry attempt %d/%d after %v delay due to error: %v", attempt, maxRetries, delay, err)
}

// LogRateLimitHit logs when rate limits are encountered
func (l *Logger) LogRateLimitHit(retryAfter time.Duration) {
	if !l.shouldLog(LogLevelWarn) {
		return
	}

	l.Warn("Rate limit hit, will retry after %v", retryAfter)
}

// LogModelUnavailable logs when a model is unavailable
func (l *Logger) LogModelUnavailable(model string, err error) {
	if !l.shouldLog(LogLevelWarn) {
		return
	}

	l.Warn("Model %s is unavailable: %v", model, err)
}

// LogConnectionTest logs the result of connection tests
func (l *Logger) LogConnectionTest(success bool, duration time.Duration, err error) {
	if success {
		l.Info("OpenRouter API connection test successful (duration: %v)", duration)
	} else {
		l.Error("OpenRouter API connection test failed (duration: %v): %v", duration, err)
	}
}

// Helper functions

// getTemperature safely gets temperature value with default
func getTemperature(temp *float32) float32 {
	if temp == nil {
		return 1.0 // Default temperature
	}
	return *temp
}

// getMaxTokens safely gets max tokens value with default
func getMaxTokens(maxTokens *int) int {
	if maxTokens == nil {
		return 0 // No limit
	}
	return *maxTokens
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// SetLogLevel sets the logging level
func (l *Logger) SetLogLevel(level LogLevel) {
	l.level = level
}

// SetMetricsEnabled enables or disables metrics logging
func (l *Logger) SetMetricsEnabled(enabled bool) {
	l.enableMetrics = enabled
}

// SetRequestLogging enables or disables request logging
func (l *Logger) SetRequestLogging(enabled bool) {
	l.enableRequestLog = enabled
}

// SetResponseLogging enables or disables response logging
func (l *Logger) SetResponseLogging(enabled bool) {
	l.enableResponseLog = enabled
}