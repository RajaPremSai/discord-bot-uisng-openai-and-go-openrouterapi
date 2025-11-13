package gpt

import (
	"testing"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

func TestValidateOpenRouterModel(t *testing.T) {
	// Test valid model
	if !validateOpenRouterModel("openai/gpt-4") {
		t.Error("Expected openai/gpt-4 to be valid")
	}
	
	// Test invalid model
	if validateOpenRouterModel("gpt-4") {
		t.Error("Expected gpt-4 to be invalid")
	}
}

func TestGetModelDisplayName(t *testing.T) {
	// Test default model
	result := getModelDisplayName("openai/gpt-4", true)
	expected := "openai/gpt-4 (Default)"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	
	// Test non-default model
	result = getModelDisplayName("anthropic/claude-3-sonnet", false)
	expected = "anthropic/claude-3-sonnet"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCommand_ModelValidation(t *testing.T) {
	client := &openrouter.Client{}
	messagesCache, _ := NewMessagesCache(10)
	ignoredChannelsCache := make(IgnoredChannelsCache)

	// Test with valid models
	models := []string{"openai/gpt-4", "anthropic/claude-3-sonnet"}
	command := Command(client, models, messagesCache, &ignoredChannelsCache)
	
	if command == nil {
		t.Fatal("Command should not be nil")
	}
	
	if command.Name != "gpt" {
		t.Errorf("Expected command name 'gpt', got %q", command.Name)
	}
	
	if command.Description != "Start conversation with AI models via OpenRouter" {
		t.Errorf("Expected updated description for OpenRouter, got %q", command.Description)
	}
}

func TestCommand_TemperatureOption(t *testing.T) {
	client := &openrouter.Client{}
	messagesCache, _ := NewMessagesCache(10)
	ignoredChannelsCache := make(IgnoredChannelsCache)
	
	command := Command(client, []string{"openai/gpt-4"}, messagesCache, &ignoredChannelsCache)
	
	// Find temperature option
	var tempOption *discord.ApplicationCommandOption
	for _, option := range command.Options {
		if option.Name == gptCommandOptionTemperature.string() {
			tempOption = option
			break
		}
	}
	
	if tempOption == nil {
		t.Fatal("Temperature option should be present")
	}
	
	// Check temperature option properties
	if tempOption.Type != discord.ApplicationCommandOptionNumber {
		t.Errorf("Expected temperature option type to be Number, got %v", tempOption.Type)
	}
	
	expectedDesc := "Sampling temperature (0.0-2.0). Lower values are more focused and deterministic"
	if tempOption.Description != expectedDesc {
		t.Errorf("Expected temperature description %q, got %q", expectedDesc, tempOption.Description)
	}
}

func TestCommand_BasicOptions(t *testing.T) {
	client := &openrouter.Client{}
	messagesCache, _ := NewMessagesCache(10)
	ignoredChannelsCache := make(IgnoredChannelsCache)
	
	command := Command(client, []string{"openai/gpt-4"}, messagesCache, &ignoredChannelsCache)
	
	// Check that basic options are present
	foundOptions := make(map[string]*discord.ApplicationCommandOption)
	for _, option := range command.Options {
		foundOptions[option.Name] = option
	}
	
	// Check prompt option
	promptOption, found := foundOptions[gptCommandOptionPrompt.string()]
	if !found {
		t.Error("Prompt option not found")
	} else {
		if promptOption.Type != discord.ApplicationCommandOptionString {
			t.Errorf("Expected prompt type String, got %v", promptOption.Type)
		}
		if !promptOption.Required {
			t.Error("Prompt should be required")
		}
		if promptOption.Description != "AI prompt for conversation" {
			t.Errorf("Expected updated prompt description, got %q", promptOption.Description)
		}
	}
}

func TestCommand_ModelFiltering(t *testing.T) {
	client := &openrouter.Client{}
	messagesCache, _ := NewMessagesCache(10)
	ignoredChannelsCache := make(IgnoredChannelsCache)

	// Test with mixed valid and invalid models
	models := []string{
		"openai/gpt-4",      // valid
		"invalid-model",     // invalid
		"anthropic/claude-3-sonnet", // valid
		"another-invalid",   // invalid
	}
	
	command := Command(client, models, messagesCache, &ignoredChannelsCache)
	
	// Find model option
	var modelOption *discord.ApplicationCommandOption
	for _, option := range command.Options {
		if option.Name == gptCommandOptionModel.string() {
			modelOption = option
			break
		}
	}
	
	// Should have model option since we have multiple valid models
	if modelOption == nil {
		t.Fatal("Expected model option to be present with multiple valid models")
	}
	
	// Should only have 2 valid models in choices
	if len(modelOption.Choices) != 2 {
		t.Errorf("Expected 2 model choices, got %d", len(modelOption.Choices))
	}
	
	// Check that only valid models are included
	validModels := make(map[string]bool)
	for _, choice := range modelOption.Choices {
		validModels[choice.Value.(string)] = true
	}
	
	if !validModels["openai/gpt-4"] {
		t.Error("Expected openai/gpt-4 to be in model choices")
	}
	
	if !validModels["anthropic/claude-3-sonnet"] {
		t.Error("Expected anthropic/claude-3-sonnet to be in model choices")
	}
	
	if validModels["invalid-model"] {
		t.Error("Did not expect invalid-model to be in model choices")
	}
}

func TestCommand_NoModelOption(t *testing.T) {
	client := &openrouter.Client{}
	messagesCache, _ := NewMessagesCache(10)
	ignoredChannelsCache := make(IgnoredChannelsCache)

	// Test with only one valid model
	models := []string{"openai/gpt-4"}
	
	command := Command(client, models, messagesCache, &ignoredChannelsCache)
	
	// Find model option
	var modelOption *discord.ApplicationCommandOption
	for _, option := range command.Options {
		if option.Name == gptCommandOptionModel.string() {
			modelOption = option
			break
		}
	}
	
	// Should not have model option since we have only one model
	if modelOption != nil {
		t.Error("Expected no model option with single model")
	}
}