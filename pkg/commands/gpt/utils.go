package gpt

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/constants"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/utils"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	gptPricePerPromptTokenGPT3Dot5Turbo0613     = 0.0000015
	gptPricePerCompletionTokenGPT3Dot5Turbo0613 = 0.000002

	gptPricePerPromptTokenGPT3Dot5Turbo16K0613     = 0.000003
	gptPricePerCompletionTokenGPT3Dot5Turbo16K0613 = 0.000004

	gptPricePerPromptTokenGPT40613     = 0.00003
	gptPricePerCompletionTokenGPT40613 = 0.00006

	gptPricePerPromptTokenGPT432K0613     = 0.00006
	gptPricePerCompletionTokenGPT432K0613 = 0.00012
)

const (
	gptTruncateLimitGPT3Dot5Turbo0301 = 3500
	gptTruncateLimitGPT40314          = 6500
	gptTruncateLimitGPT432K0314       = 30500
)

func shouldHandleMessageType(t discord.MessageType) bool {
	return t == discord.MessageTypeDefault || t == discord.MessageTypeReply
}

type chatGPTResponse struct {
	content string
	usage   openrouter.Usage
}

func sendOpenRouterRequest(client *openrouter.Client, cacheItem *MessagesCacheData) (*chatGPTResponse, error) {
	messages := cacheItem.Messages
	if cacheItem.SystemMessage != nil {
		messages = append([]openrouter.ChatCompletionMessage{*cacheItem.SystemMessage}, messages...)
	}
	req := openrouter.ChatCompletionRequest{
		Model:    cacheItem.Model,
		Messages: messages,
	}

	if cacheItem.Temperature != nil {
		req.Temperature = cacheItem.Temperature
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		req,
	)
	if err != nil {
		return nil, err
	}
	responseContent := resp.Choices[0].Message.Content
	cacheItem.Messages = append(cacheItem.Messages, openrouter.ChatCompletionMessage{
		Role:    "assistant",
		Content: responseContent,
	})
	cacheItem.TokenCount = resp.Usage.TotalTokens
	return &chatGPTResponse{
		content: responseContent,
		usage:   resp.Usage,
	}, nil
}

func sendChatGPTRequest(client *openai.Client, cacheItem *MessagesCacheData) (*chatGPTResponse, error) {
	// This function is kept for backward compatibility but should not be used with OpenRouter
	// Convert OpenRouter messages to OpenAI format for legacy support
	openaiMessages := make([]openai.ChatCompletionMessage, len(cacheItem.Messages))
	for i, msg := range cacheItem.Messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	
	var systemMessage *openai.ChatCompletionMessage
	if cacheItem.SystemMessage != nil {
		systemMessage = &openai.ChatCompletionMessage{
			Role:    cacheItem.SystemMessage.Role,
			Content: cacheItem.SystemMessage.Content,
		}
	}

	messages := openaiMessages
	if systemMessage != nil {
		messages = append([]openai.ChatCompletionMessage{*systemMessage}, messages...)
	}
	req := openai.ChatCompletionRequest{
		Model:    cacheItem.Model,
		Messages: messages,
	}

	if cacheItem.Temperature != nil {
		req.Temperature = *cacheItem.Temperature
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		req,
	)
	if err != nil {
		return nil, err
	}
	responseContent := resp.Choices[0].Message.Content
	cacheItem.Messages = append(cacheItem.Messages, openrouter.ChatCompletionMessage{
		Role:    "assistant",
		Content: responseContent,
	})
	cacheItem.TokenCount = resp.Usage.TotalTokens
	return &chatGPTResponse{
		content: responseContent,
		usage:   openrouter.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func getUrlData(client *http.Client, url string) (string, error) {
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getContentOrURLData(client *http.Client, s string) (content string, err error) {
	if utils.IsURL(s) {
		content, err = getUrlData(client, s)
	}
	return content, err
}

func parseInteractionReply(discordMessage *discord.Message) (prompt string, context string, model string, temperature *float32) {
	if discordMessage == nil || len(discordMessage.Embeds) == 0 {
		return
	}

	for _, embed := range discordMessage.Embeds {
		if embed.Description != "" {
			prompt = embed.Description
		}
		for _, field := range embed.Fields {
			switch field.Name {
			case gptCommandOptionPrompt.humanReadableString():
				prompt = field.Value
			case gptCommandOptionContext.humanReadableString():
				if context == "" {
					// file context always gets precedence
					context = field.Value
				}
			case gptCommandOptionContextFile.humanReadableString():
				context = field.Value
			case gptCommandOptionModel.humanReadableString():
				model = field.Value
			case gptCommandOptionTemperature.humanReadableString():
				parsedValue, err := strconv.ParseFloat(field.Value, 32)
				if err != nil {
					log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to parse temperature value from the message with the error: %v\n", discordMessage.GuildID, discordMessage.ChannelID, discordMessage.ID, err)
					continue
				}
				temp := float32(parsedValue)
				temperature = &temp
			}
		}
	}

	return
}

func modelTruncateLimit(model string) *int {
	// Extract base model for OpenRouter format
	baseModel := extractBaseModel(model)
	
	var truncateLimit int
	switch baseModel {
	case "gpt-3.5-turbo":
		truncateLimit = gptTruncateLimitGPT3Dot5Turbo0301
	case "gpt-4":
		truncateLimit = gptTruncateLimitGPT40314
	case "gpt-4-32k":
		truncateLimit = gptTruncateLimitGPT432K0314
	default:
		//to be implemented
		return nil
	}
	return &truncateLimit
}

func generateCost(usage openai.Usage, model string) string {
	var cost float64

	switch model {
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo0301, openai.GPT3Dot5Turbo0613:
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT3Dot5Turbo0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT3Dot5Turbo0613
	case openai.GPT3Dot5Turbo16K, openai.GPT3Dot5Turbo16K0613:
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT3Dot5Turbo16K0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT3Dot5Turbo16K0613
	case openai.GPT4, openai.GPT40314, openai.GPT40613:
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT40613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT40613
	case openai.GPT432K, openai.GPT432K0314, openai.GPT432K0613:
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT432K0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT432K0613
	default:
		// to be implemented
		return ""
	}

	return fmt.Sprintf("\nLLM Cost: $%f", cost)
}

func generateOpenRouterCost(usage openrouter.Usage, model string) string {
	// OpenRouter provides cost information directly in the response
	if usage.TotalCost > 0 {
		return fmt.Sprintf("\nLLM Cost: $%.6f", usage.TotalCost)
	}
	
	// Fallback to estimated cost based on model type for OpenRouter models
	var cost float64
	
	// Extract base model from OpenRouter format (e.g., "openai/gpt-4" -> "gpt-4")
	baseModel := model
	if strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		if len(parts) > 1 {
			baseModel = parts[1]
		}
	}
	
	switch baseModel {
	case "gpt-3.5-turbo":
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT3Dot5Turbo0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT3Dot5Turbo0613
	case "gpt-4":
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT40613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT40613
	default:
		// For unknown models, don't show cost estimation
		return ""
	}

	return fmt.Sprintf("\nEstimated Cost: $%.6f", cost)
}

func adjustMessageTokens(cacheItem *MessagesCacheData) {
	truncateLimit := modelTruncateLimit(cacheItem.Model)
	if truncateLimit == nil {
		return
	}

	for cacheItem.TokenCount > *truncateLimit {
		message := cacheItem.Messages[0]
		cacheItem.Messages = cacheItem.Messages[1:]
		removedTokens := countOpenRouterMessageTokens(message, cacheItem.Model)
		if removedTokens == nil {
			return
		}
		cacheItem.TokenCount -= *removedTokens
	}
}

func isCacheItemWithinTruncateLimit(cacheItem *MessagesCacheData) (ok bool, count int) {
	truncateLimit := modelTruncateLimit(cacheItem.Model)
	if truncateLimit == nil {
		return true, 0
	}

	tokens := countAllOpenRouterMessagesTokens(cacheItem.SystemMessage, cacheItem.Messages, cacheItem.Model)
	if tokens == nil {
		return true, 0
	}
	cacheItem.TokenCount = *tokens

	return *tokens <= *truncateLimit, *tokens
}

func generateThreadTitleBasedOnInitialPrompt(ctx *bot.Context, client *openrouter.Client, threadID string, messages []openrouter.ChatCompletionChoice) {
	conversation := make([]map[string]string, len(messages))
	for i, msg := range messages {
		conversation[i] = map[string]string{
			"role":    msg.Message.Role,
			"content": msg.Message.Content,
		}
	}
	var conversationTextBuilder strings.Builder
	for _, msg := range conversation {
		conversationTextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg["role"], msg["content"]))
	}
	conversationText := conversationTextBuilder.String()

	prompt := fmt.Sprintf("%s\nGenerate a short and concise title summarizing the conversation in the same language. The title must not contain any quotes. The title should be no longer than 60 characters:", conversationText)

	// Use chat completion instead of completion for OpenRouter
	resp, err := client.CreateChatCompletion(context.Background(), openrouter.ChatCompletionRequest{
		Model: "openai/gpt-3.5-turbo", // Use a reliable model for title generation
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: func() *float32 { t := float32(0.5); return &t }(),
		MaxTokens:   func() *int { t := 75; return &t }(),
	})
	if err != nil {
		log.Printf("[GID: %s, threadID: %s] Failed to generate thread title with the error: %v\n", ctx.Interaction.GuildID, threadID, err)
		return
	}

	title := resp.Choices[0].Message.Content
	if len(title) > 60 {
		title = title[:60]
	}

	_, err = ctx.Session.ChannelEditComplex(threadID, &discord.ChannelEdit{
		Name: title,
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to update thread title with the error: %v\n", ctx.Interaction.GuildID, threadID, err)
	}
}

func attachUsageInfo(s *discord.Session, m *discord.Message, usage openrouter.Usage, model string) {
	var extraInfo string
	if usage.TotalCost > 0 {
		// OpenRouter provides cost information directly
		extraInfo = fmt.Sprintf("Completion Tokens: %d, Total: %d, Cost: $%.6f", usage.CompletionTokens, usage.TotalTokens, usage.TotalCost)
	} else {
		// Fallback to token count only if cost is not available
		extraInfo = fmt.Sprintf("Completion Tokens: %d, Total: %d%s", usage.CompletionTokens, usage.TotalTokens, generateOpenRouterCost(usage, model))
	}

	utils.DiscordChannelMessageEdit(s, m.ID, m.ChannelID, nil, []*discord.MessageEmbed{
		{
			Footer: &discord.MessageEmbedFooter{
				Text:    extraInfo,
				IconURL: constants.OpenRouterIconURL,
			},
		},
	})
}
