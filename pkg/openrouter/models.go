package openrouter

import (
	"encoding/json"
	"fmt"
)

// ChatCompletionRequest represents a chat completion request to OpenRouter
type ChatCompletionRequest struct {
	Model            string                    `json:"model"`
	Messages         []ChatCompletionMessage   `json:"messages"`
	Temperature      *float32                  `json:"temperature,omitempty"`
	MaxTokens        *int                      `json:"max_tokens,omitempty"`
	TopP             *float32                  `json:"top_p,omitempty"`
	FrequencyPenalty *float32                  `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32                  `json:"presence_penalty,omitempty"`
	Stream           bool                      `json:"stream"`
	Stop             []string                  `json:"stop,omitempty"`
	User             string                    `json:"user,omitempty"`
}

// ChatCompletionMessage represents a message in a chat completion
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatCompletionResponse represents the response from OpenRouter chat completion
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// ChatCompletionChoice represents a choice in the chat completion response
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
	LogProbs     *LogProbs             `json:"logprobs,omitempty"`
}

// LogProbs represents log probabilities for tokens
type LogProbs struct {
	Tokens        []string             `json:"tokens"`
	TokenLogProbs []float32            `json:"token_logprobs"`
	TopLogProbs   []map[string]float32 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// ImageRequest represents an image generation request to OpenRouter
type ImageRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
}

// ImageResponse represents the response from OpenRouter image generation
type ImageResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// ImageData represents individual image data in the response
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	PromptCost       float64 `json:"prompt_cost,omitempty"`
	CompletionCost   float64 `json:"completion_cost,omitempty"`
	TotalCost        float64 `json:"total_cost,omitempty"`
}

// ErrorResponse represents an error response from OpenRouter
type ErrorResponse struct {
	ErrorDetail ErrorDetail `json:"error"`
}

// ErrorDetail represents the detailed error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   string      `json:"param,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface for ErrorResponse
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("OpenRouter API error: %s (code: %s, type: %s)", e.ErrorDetail.Message, e.ErrorDetail.Code, e.ErrorDetail.Type)
}

// ModelsResponse represents the response from the models endpoint
type ModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

// Model represents an available model from OpenRouter
type Model struct {
	ID         string      `json:"id"`
	Object     string      `json:"object"`
	Created    int64       `json:"created"`
	OwnedBy    string      `json:"owned_by"`
	Permission []ModelPerm `json:"permission,omitempty"`
	Root       string      `json:"root,omitempty"`
	Parent     string      `json:"parent,omitempty"`
}

// ModelPerm represents model permissions
type ModelPerm struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int64       `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogProbs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []StreamChoice       `json:"choices"`
	Usage   *Usage               `json:"usage,omitempty"`
}

// StreamChoice represents a choice in a streaming response
type StreamChoice struct {
	Index        int                   `json:"index"`
	Delta        ChatCompletionMessage `json:"delta"`
	FinishReason string                `json:"finish_reason"`
}

// Validate validates the ChatCompletionRequest
func (r *ChatCompletionRequest) Validate() error {
	if r.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}
	for i, msg := range r.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}
	}
	return nil
}

// Validate validates the ImageRequest
func (r *ImageRequest) Validate() error {
	if r.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if r.Model == "" {
		return fmt.Errorf("model is required")
	}
	if r.N < 0 {
		return fmt.Errorf("n must be non-negative")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for ChatCompletionRequest
func (r *ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	type Alias ChatCompletionRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for ChatCompletionRequest
func (r *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	type Alias ChatCompletionRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// MarshalJSON implements custom JSON marshaling for ImageRequest
func (r *ImageRequest) MarshalJSON() ([]byte, error) {
	type Alias ImageRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for ImageRequest
func (r *ImageRequest) UnmarshalJSON(data []byte) error {
	type Alias ImageRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// Helper functions for creating pointers to primitive types

// IntPtr returns a pointer to the given int value
func IntPtr(v int) *int {
	return &v
}

// Float32Ptr returns a pointer to the given float32 value
func Float32Ptr(v float32) *float32 {
	return &v
}

// StringPtr returns a pointer to the given string value
func StringPtr(v string) *string {
	return &v
}

// BoolPtr returns a pointer to the given bool value
func BoolPtr(v bool) *bool {
	return &v
}