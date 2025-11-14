package gpt

import (
	"strings"
	
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

func countMessageTokens(message openai.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(model)
	if !ok {
		return nil
	}
	enc, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}
	tokens := _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	return &tokens
}

func countOpenRouterMessageTokens(message openrouter.ChatCompletionMessage, model string) *int {
	// Convert OpenRouter message to OpenAI format for token counting
	openaiMessage := openai.ChatCompletionMessage{
		Role:    message.Role,
		Content: message.Content,
		Name:    message.Name,
	}
	return countMessageTokens(openaiMessage, extractBaseModel(model))
}

func countMessagesTokens(messages []openai.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(model)
	if !ok {
		return nil
	}

	enc, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}

	tokens := 0
	for _, message := range messages {
		tokens += _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	}
	tokens += 3

	return &tokens
}

func _countMessageTokens(enc tokenizer.Codec, tokensPerMessage int, tokensPerName int, message openai.ChatCompletionMessage) int {
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

func countAllMessagesTokens(systemMessage *openai.ChatCompletionMessage, messages []openai.ChatCompletionMessage, model string) *int {
	if systemMessage != nil {
		messages = append(messages, *systemMessage)
	}
	return countMessagesTokens(messages, model)
}

func countAllOpenRouterMessagesTokens(systemMessage *openrouter.ChatCompletionMessage, messages []openrouter.ChatCompletionMessage, model string) *int {
	// Convert OpenRouter messages to OpenAI format for token counting
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}
	
	var openaiSystemMessage *openai.ChatCompletionMessage
	if systemMessage != nil {
		openaiSystemMessage = &openai.ChatCompletionMessage{
			Role:    systemMessage.Role,
			Content: systemMessage.Content,
			Name:    systemMessage.Name,
		}
	}
	
	return countAllMessagesTokens(openaiSystemMessage, openaiMessages, extractBaseModel(model))
}

func _tokensConfiguration(model string) (ok bool, tokensPerMessage int, tokensPerName int) {
	ok = true

	switch model {
	case openai.GPT3Dot5Turbo0301:
		tokensPerMessage = 4
		tokensPerName = -1
	case openai.GPT3Dot5Turbo,
		openai.GPT3Dot5Turbo0613,
		openai.GPT3Dot5Turbo16K,
		openai.GPT3Dot5Turbo16K0613,
		openai.GPT4,
		openai.GPT40314,
		openai.GPT40613,
		openai.GPT432K0314,
		openai.GPT432K0613:
		tokensPerMessage = 3
		tokensPerName = 1
	default:
		// For unknown models (including non-OpenAI models from OpenRouter),
		// use GPT-4 configuration as a reasonable default
		tokensPerMessage = 3
		tokensPerName = 1
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

