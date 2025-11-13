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
}

// ClientConfig holds configuration for the OpenRouter client
type ClientConfig struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	SiteURL    string
	SiteName   string
}

// NewClient creates a new OpenRouter API client
func NewClient(apiKey string) *Client {
	return NewClientWithConfig(ClientConfig{
		APIKey:  apiKey,
		BaseURL: DefaultBaseURL,
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

	return &Client{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		httpClient: httpClient,
		siteURL:    config.SiteURL,
		siteName:   config.SiteName,
	}
}

// SetSiteInfo sets the site URL and name for OpenRouter headers
func (c *Client) SetSiteInfo(siteURL, siteName string) {
	c.siteURL = siteURL
	c.siteName = siteName
}

// buildRequest creates an HTTP request with proper OpenRouter headers
func (c *Client) buildRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
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

	return req, nil
}

// doRequest executes an HTTP request and handles the response
func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return &errorResp
	}

	// Parse successful response
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// CreateChatCompletion creates a chat completion using the OpenRouter API
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	httpReq, err := c.buildRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	var resp ChatCompletionResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateImage creates an image using the OpenRouter API
func (c *Client) CreateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	httpReq, err := c.buildRequest(ctx, "POST", "/images/generations", req)
	if err != nil {
		return nil, err
	}

	var resp ImageResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
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
	httpReq, err := c.buildRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return err
	}

	return c.doRequest(httpReq, nil)
}