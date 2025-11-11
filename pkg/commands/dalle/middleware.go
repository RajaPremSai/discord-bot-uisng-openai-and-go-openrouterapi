package dalle

import (
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/sashabaranov/go-openai"
)

func imageInteractionResponseMiddleware(ctx *bot.Context)
func imageModerationMiddleware(ctx *bot.Context, client *openai.Client)
