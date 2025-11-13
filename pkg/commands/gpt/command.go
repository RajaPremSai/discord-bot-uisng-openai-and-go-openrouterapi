package gpt

import (
	"strings"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

const commandName = "gpt"

var gptDefaultModel = "openai/gpt-3.5-turbo"

// validateOpenRouterModel validates that the model name follows OpenRouter format (provider/model)
func validateOpenRouterModel(model string) bool {
	return strings.Contains(model, "/") && len(strings.Split(model, "/")) == 2
}

// getModelDisplayName returns a user-friendly display name for the model
func getModelDisplayName(model string, isDefault bool) string {
	name := model
	if isDefault {
		name += " (Default)"
	}
	return name
}

func Command(client *openrouter.Client, completionModels []string, messagesCache *MessagesCache, ignoredChannelsCache *IgnoredChannelsCache) *bot.Command {
	temperatureOptionMinValue := 0.0
	temperatureOptionMaxValue := 2.0
	
	opts := []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionPrompt.string(),
			Description: "AI prompt for conversation",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionContext.string(),
			Description: "Sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
		{
			Type:        discord.ApplicationCommandOptionAttachment,
			Name:        gptCommandOptionContextFile.string(),
			Description: "File that sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
	}
	
	// Validate and filter OpenRouter models
	validModels := make([]string, 0, len(completionModels))
	for _, model := range completionModels {
		if validateOpenRouterModel(model) {
			validModels = append(validModels, model)
		}
	}
	
	numberOfModels := len(validModels)
	if numberOfModels > 0 {
		gptDefaultModel = validModels[0]
	}
	
	// Add model selection option if multiple models are available
	if numberOfModels > 1 {
		var modelChoices []*discord.ApplicationCommandOptionChoice
		for i, model := range validModels {
			modelChoices = append(modelChoices, &discord.ApplicationCommandOptionChoice{
				Name:  getModelDisplayName(model, i == 0),
				Value: model,
			})
		}
		opts = append(opts, &discord.ApplicationCommandOption{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionModel.string(),
			Description: "AI model to use (OpenRouter format: provider/model)",
			Required:    false,
			Choices:     modelChoices,
		})
	}
	
	// Add temperature option with OpenRouter-compatible range
	opts = append(opts, &discord.ApplicationCommandOption{
		Type:        discord.ApplicationCommandOptionNumber,
		Name:        gptCommandOptionTemperature.string(),
		Description: "Sampling temperature (0.0-2.0). Lower values are more focused and deterministic",
		MinValue:    &temperatureOptionMinValue,
		MaxValue:    temperatureOptionMaxValue,
		Required:    false,
	})
	
	return &bot.Command{
		Name:        commandName,
		Description: "Start conversation with AI models via OpenRouter",
		Options:     opts,
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			chatGPTHandler(ctx, client, messagesCache)
		}),
		MessageHandler: bot.MessageHandlerFunc(func(ctx *bot.MessageContext) {
			chatGPTMessageHandler(ctx, client, messagesCache, ignoredChannelsCache)
		}),
	}
}
