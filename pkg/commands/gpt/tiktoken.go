package gpt

import (
	"strings"
	
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	"github.com/tiktoken-go/tokenizer"
)



func countOpenRouterMessageTokens(message openrouter.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(extractBaseModel(model))
	if !ok {
		return nil
	}
	enc, err := tokenizer.ForModel(tokenizer.Model(extractBaseModel(model)))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}
	tokens := _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	return &tokens
}



func _countMessageTokens(enc tokenizer.Codec, tokensPerMessage int, tokensPerName int, message openrouter.ChatCompletionMessage) int {
	tokens := tokensPerMessage
	contentIds, _, _ := enc.Encode(message.Content)
	roleIds, _, _ := enc.Encode(message.Role)
	tokens += len(contentIds)
	tokens += len(roleIds)
	if message.Name != "" {
		tokens += tokensPerName
		nameIds, _, _ := enc.Encode(message.Name)
		tokens += len(nameIds)
	}
	return tokens
}



func countAllOpenRouterMessagesTokens(systemMessage *openrouter.ChatCompletionMessage, messages []openrouter.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(extractBaseModel(model))
	if !ok {
		return nil
	}

	enc, err := tokenizer.ForModel(tokenizer.Model(extractBaseModel(model)))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}

	tokens := 0
	for _, message := range messages {
		tokens += _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	}
	if systemMessage != nil {
		tokens += _countMessageTokens(enc, tokensPerMessage, tokensPerName, *systemMessage)
	}
	tokens += 3

	return &tokens
}

func _tokensConfiguration(model string) (ok bool, tokensPerMessage int, tokensPerName int) {
	ok = true

	switch model {
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4
		tokensPerName = -1
	case "gpt-3.5-turbo", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-4", "gpt-4-0314", "gpt-4-0613", "gpt-4-32k-0314", "gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	default:
		// For unknown models (including OpenRouter models),
		// check if it's a known model family and use appropriate defaults
		if strings.Contains(model, "gpt-4") || strings.Contains(model, "claude") {
			tokensPerMessage = 3
			tokensPerName = 1
		} else if strings.Contains(model, "gpt-3.5") {
			tokensPerMessage = 3
			tokensPerName = 1
		} else {
			// Use GPT-4 configuration as a reasonable default for unknown models
			tokensPerMessage = 3
			tokensPerName = 1
		}
		ok = true
	}
	return
}

// extractBaseModel extracts the base model name from OpenRouter format
// e.g., "openai/gpt-4" -> "gpt-4", "anthropic/claude-3-sonnet" -> "claude-3-sonnet"
func extractBaseModel(model string) string {
	if strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return model
}

