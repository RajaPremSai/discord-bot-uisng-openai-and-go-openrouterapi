package openrouter

import "context"

// ChatCompletionClient defines the interface for chat completion operations
type ChatCompletionClient interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

// ImageGenerationClient defines the interface for image generation operations
type ImageGenerationClient interface {
	CreateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error)
}

// OpenRouterClient combines all OpenRouter API operations
type OpenRouterClient interface {
	ChatCompletionClient
	ImageGenerationClient
}

// Ensure Client implements OpenRouterClient interface
var _ OpenRouterClient = (*Client)(nil)