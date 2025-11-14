package gpt

import (
	"strings"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	"github.com/sashabaranov/go-openai"
)

const discordMaxMessageLength = 2000

func splitMessage(message string) []string {
	if len(message) <= discordMaxMessageLength {
		// the message is short enough to be sent as is
		return []string{message}
	}

	// split the message by whitespace
	words := strings.Fields(message)
	var messageParts []string
	currentMessage := ""
	for _, word := range words {
		if len(currentMessage)+len(word)+1 > discordMaxMessageLength {
			// start a new message if adding the current word exceeds the maximum length
			messageParts = append(messageParts, currentMessage)
			currentMessage = word + " "
		} else {
			// add the current word to the current message
			currentMessage += word + " "
		}
	}
	// add the last message to the list of message parts
	messageParts = append(messageParts, currentMessage)

	return messageParts
}

func reverseMessages(messages *[]openai.ChatCompletionMessage) {
	length := len(*messages)
	for i := 0; i < length/2; i++ {
		(*messages)[i], (*messages)[length-i-1] = (*messages)[length-i-1], (*messages)[i]
	}
}

func reverseOpenRouterMessages(messages *[]openrouter.ChatCompletionMessage) {
	length := len(*messages)
	for i := 0; i < length/2; i++ {
		(*messages)[i], (*messages)[length-i-1] = (*messages)[length-i-1], (*messages)[i]
	}
}

// normalizeOpenRouterModelName returns a user-friendly model name for display
// e.g., "openai/gpt-4" -> "GPT-4", "anthropic/claude-3-sonnet" -> "Claude-3-Sonnet"
func normalizeOpenRouterModelName(model string) string {
	if strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		if len(parts) > 1 {
			baseModel := parts[1]
			// Capitalize common model names for better display
			switch {
			case strings.HasPrefix(baseModel, "gpt"):
				return strings.ToUpper(baseModel)
			case strings.HasPrefix(baseModel, "claude"):
				// Manually capitalize first letter of each word for Claude models
				words := strings.Split(baseModel, "-")
				for i, word := range words {
					if len(word) > 0 {
						words[i] = strings.ToUpper(word[:1]) + word[1:]
					}
				}
				return strings.Join(words, "-")
			default:
				return baseModel
			}
		}
	}
	return model
}
