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
		Description: "Generate creative images from textual description using OpenRouter AI models",
		Options: []*discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionPrompt.String(),
				Description: "A text description of the desired image",
				Required:    true,
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionModel.String(),
				Description: "The AI model to use for image generation",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  "DALL-E 2 (Default)",
						Value: "openai/dall-e-2",
					},
					{
						Name:  "DALL-E 3 (Higher Quality)",
						Value: "openai/dall-e-3",
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionSize.String(),
				Description: "The size of the generated images",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  "256x256 (DALL-E 2 only)",
						Value: "256x256",
					},
					{
						Name:  "512x512 (DALL-E 2 only)",
						Value: "512x512",
					},
					{
						Name:  "1024x1024 (Default)",
						Value: "1024x1024",
					},
					{
						Name:  "1024x1792 (DALL-E 3 only)",
						Value: "1024x1792",
					},
					{
						Name:  "1792x1024 (DALL-E 3 only)",
						Value: "1792x1024",
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionInteger,
				Name:        imageCommandOptionNumber.String(),
				Description: "The number of images to generate (default 1, max 4 for DALL-E 2, max 1 for DALL-E 3)",
				MinValue:    &numberOptionMinValue,
				MaxValue:    4,
				Required:    false,
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionQuality.String(),
				Description: "Image quality (DALL-E 3 only)",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  "Standard (Default)",
						Value: "standard",
					},
					{
						Name:  "HD (Higher Quality)",
						Value: "hd",
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionStyle.String(),
				Description: "Image style (DALL-E 3 only)",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  "Vivid (Default)",
						Value: "vivid",
					},
					{
						Name:  "Natural",
						Value: "natural",
					},
				},
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
