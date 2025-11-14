package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the default OpenRouter API base URL
	DefaultBaseURL = "https://openrouter.ai/api/v1"
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client represents an OpenRouter API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	siteURL    string
	siteName   string
	logger     *Logger
}

// ClientConfig holds configuration for the OpenRouter client
type ClientConfig struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	SiteURL    string
	SiteName   string
	Logger     *Logger
}

// NewClient creates a new OpenRouter API client
func NewClient(apiKey string) *Client {
	return NewClientWithConfig(ClientConfig{
		APIKey:  apiKey,
		BaseURL: DefaultBaseURL,
		Logger:  DefaultLogger(),
	})
}

// NewClientWithConfig creates a new OpenRouter API client with custom configuration
func NewClientWithConfig(config ClientConfig) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	logger := config.Logger
	if logger == nil {
		logger = DefaultLogger()
	}

	return &Client{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		httpClient: httpClient,
		siteURL:    config.SiteURL,
		siteName:   config.SiteName,
		logger:     logger,
	}
}

// SetSiteInfo sets the site URL and name for OpenRouter headers
func (c *Client) SetSiteInfo(siteURL, siteName string) {
	c.siteURL = siteURL
	c.siteName = siteName
}

// SetLogger sets the logger for the client
func (c *Client) SetLogger(logger *Logger) {
	c.logger = logger
}

// GetLogger returns the client's logger
func (c *Client) GetLogger() *Logger {
	return c.logger
}

// WithRetry executes a function with retry logic and logging
func (c *Client) WithRetry(ctx context.Context, config *RetryConfig, fn RetryableFunc) error {
	return WithRetry(ctx, config, c.logger, fn)
}

// buildRequest creates an HTTP request with proper OpenRouter headers
func (c *Client) buildRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.logger.LogError(err, "Marshaling request body")
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		c.logger.LogError(err, "Creating HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Set optional OpenRouter-specific headers
	if c.siteURL != "" {
		req.Header.Set("HTTP-Referer", c.siteURL)
	}
	if c.siteName != "" {
		req.Header.Set("X-Title", c.siteName)
	}

	// Log the request
	c.logger.LogRequest(req, body)

	return req, nil
}

// doRequest executes an HTTP request and handles the response
func (c *Client) doRequest(req *http.Request, result interface{}) error {
	startTime := time.Now()
	
	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		// Log network error
		c.logger.LogError(WrapNetworkError(err), fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path))
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.LogError(err, "Reading response body")
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response details
	var responseBody interface{}
	if len(body) > 0 {
		json.Unmarshal(body, &responseBody) // Best effort, ignore errors
	}
	c.logger.LogResponse(resp.StatusCode, resp.Header, responseBody, duration)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			httpErr := fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			c.logger.LogError(httpErr, fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path))
			return httpErr
		}
		
		// Create structured error and log it
		orErr := ParseError(resp, body)
		c.logger.LogError(orErr, fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path))
		return orErr
	}

	// Parse successful response
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			c.logger.LogError(err, "Unmarshaling response")
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// CreateChatCompletion creates a chat completion using the OpenRouter API
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	startTime := time.Now()
	
	if err := req.Validate(); err != nil {
		c.logger.LogError(err, "Chat completion request validation")
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	httpReq, err := c.buildRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	var resp ChatCompletionResponse
	err = c.doRequest(httpReq, &resp)
	duration := time.Since(startTime)
	
	// Log chat completion specific metrics
	c.logger.LogChatCompletion(req, &resp, duration, err)
	
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateImage creates an image using the OpenRouter API
func (c *Client) CreateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	startTime := time.Now()
	
	if err := req.Validate(); err != nil {
		c.logger.LogError(err, "Image generation request validation")
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	httpReq, err := c.buildRequest(ctx, "POST", "/images/generations", req)
	if err != nil {
		return nil, err
	}

	var resp ImageResponse
	err = c.doRequest(httpReq, &resp)
	duration := time.Since(startTime)
	
	// Log image generation specific metrics
	c.logger.LogImageGeneration(req, &resp, duration, err)
	
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListModels retrieves the list of available models from OpenRouter
func (c *Client) ListModels(ctx context.Context) (*ModelsResponse, error) {
	httpReq, err := c.buildRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return nil, err
	}

	var resp ModelsResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetModel retrieves information about a specific model
func (c *Client) GetModel(ctx context.Context, modelID string) (*Model, error) {
	endpoint := fmt.Sprintf("/models/%s", modelID)
	httpReq, err := c.buildRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var model Model
	if err := c.doRequest(httpReq, &model); err != nil {
		return nil, err
	}

	return &model, nil
}

// Ping tests the connection to OpenRouter API
func (c *Client) Ping(ctx context.Context) error {
	startTime := time.Now()
	
	httpReq, err := c.buildRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return err
	}

	err = c.doRequest(httpReq, nil)
	duration := time.Since(startTime)
	
	// Log connection test result
	c.logger.LogConnectionTest(err == nil, duration, err)
	
	return err
}