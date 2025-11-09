package gpt

import (
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/sashabaranov/go-openai"
)

const (
	gptDiscordThreadAutoArchivewDurationMinutes = 60
	gptInteractionEmbedColor                    = 0x000000
	gptPendingMessage                           = "⌛ Wait a moment, please..."
	gptContextOptionMaxLength                   = 1024
)

func chatGPTHandler(ctx *bot.Context, client *openai.Client, messagesCache *MessagesCache) {
}
