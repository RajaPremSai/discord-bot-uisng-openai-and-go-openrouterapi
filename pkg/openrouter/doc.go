// Package openrouter provides a client for interacting with the OpenRouter API.
//
// OpenRouter is a unified interface to multiple AI models, providing access to
// various language models and image generation models through a single API.
//
// This package implements:
//   - Chat completion functionality compatible with OpenAI's chat API
//   - Image generation functionality for DALL-E and other image models
//   - Proper error handling and response parsing for OpenRouter-specific responses
//   - Authentication and request formatting for OpenRouter API requirements
//
// Basic usage:
//
//	config := openrouter.ClientConfig{
//		APIKey:   "sk-or-v1-...",
//		BaseURL:  "https://openrouter.ai/api/v1",
//		SiteURL:  "https://yoursite.com",
//		SiteName: "Your App Name",
//	}
//	
//	client := openrouter.NewClient(config)
//	
//	req := openrouter.ChatCompletionRequest{
//		Model: "openai/gpt-4",
//		Messages: []openrouter.ChatCompletionMessage{
//			{Role: "user", Content: "Hello, world!"},
//		},
//	}
//	
//	resp, err := client.CreateChatCompletion(context.Background(), req)
package openrouter