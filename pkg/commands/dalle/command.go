package dalle

import (
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

const commandName = "dalle"

func Command(client *openrouter.Client, imageModel string) *bot.Command {
	numberOptionMinValue := 1.0
	return &bot.Command{
		Name:        commandName,
		Description: "Generate creative imgs from txt desc using OPENAI Dalle 2",
		Options: []*discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionPrompt.String(),
				Description: "A txt desc of the desired image",
				Required:    true,
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionSize.String(),
				Description: "The size of the generated images",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  "256x256 (Default)",
						Value: "256x256",
					},
					{
						Name:  "512x512",
						Value: "512x512",
					},
					{
						Name:  "1024x1024",
						Value: "1024x1024",
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionInteger,
				Name:        imageCommandOptionNumber.String(),
				Description: "The number of images to generate (default 1, max 4)",
				MinValue:    &numberOptionMinValue,
				MaxValue:    4,
				Required:    false,
			},
		},
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			imageHandler(ctx, client, imageModel)
		}),
		Middlewares: []bot.Handler{
			bot.HandlerFunc(imageInteractionResponseMiddleware),
			bot.HandlerFunc(func(ctx *bot.Context) {
				imageModerationMiddleware(ctx, client)
			}),
		},
	}
}
