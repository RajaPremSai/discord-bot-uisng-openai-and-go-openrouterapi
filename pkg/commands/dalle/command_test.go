package dalle

import (
	"testing"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

func TestCommand(t *testing.T) {
	// Create a mock OpenRouter client
	client := openrouter.NewClient("test-api-key")
	imageModel := "openai/dall-e-2"

	// Create the command
	cmd := Command(client, imageModel)

	// Test basic command properties
	if cmd.Name != commandName {
		t.Errorf("Expected command name %s, got %s", commandName, cmd.Name)
	}

	if cmd.Description == "" {
		t.Error("Command description should not be empty")
	}

	// Test that all required options are present
	expectedOptions := map[string]bool{
		"prompt":  false,
		"model":   false,
		"size":    false,
		"number":  false,
		"quality": false,
		"style":   false,
	}

	for _, option := range cmd.Options {
		if _, exists := expectedOptions[option.Name]; exists {
			expectedOptions[option.Name] = true
		}
	}

	// Check that all expected options were found
	for optionName, found := range expectedOptions {
		if !found {
			t.Errorf("Expected option %s not found in command", optionName)
		}
	}

	// Test prompt option (required)
	promptOption := findOptionByName(cmd.Options, "prompt")
	if promptOption == nil {
		t.Fatal("Prompt option not found")
	}
	if !promptOption.Required {
		t.Error("Prompt option should be required")
	}
	if promptOption.Type != discord.ApplicationCommandOptionString {
		t.Error("Prompt option should be of type String")
	}

	// Test model option (optional with choices)
	modelOption := findOptionByName(cmd.Options, "model")
	if modelOption == nil {
		t.Fatal("Model option not found")
	}
	if modelOption.Required {
		t.Error("Model option should be optional")
	}
	if len(modelOption.Choices) == 0 {
		t.Error("Model option should have choices")
	}

	// Verify model choices
	expectedModels := []string{"openai/dall-e-2", "openai/dall-e-3"}
	modelChoices := make(map[string]bool)
	for _, choice := range modelOption.Choices {
		if value, ok := choice.Value.(string); ok {
			modelChoices[value] = true
		}
	}
	for _, expectedModel := range expectedModels {
		if !modelChoices[expectedModel] {
			t.Errorf("Expected model choice %s not found", expectedModel)
		}
	}

	// Test size option (optional with choices)
	sizeOption := findOptionByName(cmd.Options, "size")
	if sizeOption == nil {
		t.Fatal("Size option not found")
	}
	if sizeOption.Required {
		t.Error("Size option should be optional")
	}
	if len(sizeOption.Choices) == 0 {
		t.Error("Size option should have choices")
	}

	// Verify size choices include both DALL-E 2 and DALL-E 3 sizes
	expectedSizes := []string{"256x256", "512x512", "1024x1024", "1024x1792", "1792x1024"}
	sizeChoices := make(map[string]bool)
	for _, choice := range sizeOption.Choices {
		if value, ok := choice.Value.(string); ok {
			sizeChoices[value] = true
		}
	}
	for _, expectedSize := range expectedSizes {
		if !sizeChoices[expectedSize] {
			t.Errorf("Expected size choice %s not found", expectedSize)
		}
	}

	// Test number option (optional with min/max values)
	numberOption := findOptionByName(cmd.Options, "number")
	if numberOption == nil {
		t.Fatal("Number option not found")
	}
	if numberOption.Required {
		t.Error("Number option should be optional")
	}
	if numberOption.Type != discord.ApplicationCommandOptionInteger {
		t.Error("Number option should be of type Integer")
	}
	if numberOption.MinValue == nil || *numberOption.MinValue != 1.0 {
		t.Error("Number option should have minimum value of 1")
	}
	if numberOption.MaxValue != 4 {
		t.Error("Number option should have maximum value of 4")
	}

	// Test quality option (optional with choices)
	qualityOption := findOptionByName(cmd.Options, "quality")
	if qualityOption == nil {
		t.Fatal("Quality option not found")
	}
	if qualityOption.Required {
		t.Error("Quality option should be optional")
	}
	if len(qualityOption.Choices) == 0 {
		t.Error("Quality option should have choices")
	}

	// Verify quality choices
	expectedQualities := []string{"standard", "hd"}
	qualityChoices := make(map[string]bool)
	for _, choice := range qualityOption.Choices {
		if value, ok := choice.Value.(string); ok {
			qualityChoices[value] = true
		}
	}
	for _, expectedQuality := range expectedQualities {
		if !qualityChoices[expectedQuality] {
			t.Errorf("Expected quality choice %s not found", expectedQuality)
		}
	}

	// Test style option (optional with choices)
	styleOption := findOptionByName(cmd.Options, "style")
	if styleOption == nil {
		t.Fatal("Style option not found")
	}
	if styleOption.Required {
		t.Error("Style option should be optional")
	}
	if len(styleOption.Choices) == 0 {
		t.Error("Style option should have choices")
	}

	// Verify style choices
	expectedStyles := []string{"vivid", "natural"}
	styleChoices := make(map[string]bool)
	for _, choice := range styleOption.Choices {
		if value, ok := choice.Value.(string); ok {
			styleChoices[value] = true
		}
	}
	for _, expectedStyle := range expectedStyles {
		if !styleChoices[expectedStyle] {
			t.Errorf("Expected style choice %s not found", expectedStyle)
		}
	}
}

func TestCommandOptionTypes(t *testing.T) {
	tests := []struct {
		option   imageCommandOptionType
		expected string
	}{
		{imageCommandOptionPrompt, "prompt"},
		{imageCommandOptionModel, "model"},
		{imageCommandOptionSize, "size"},
		{imageCommandOptionNumber, "number"},
		{imageCommandOptionQuality, "quality"},
		{imageCommandOptionStyle, "style"},
	}

	for _, test := range tests {
		result := test.option.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s for option type %d", test.expected, result, test.option)
		}
	}
}

func TestCommandOptionTypeUnknown(t *testing.T) {
	unknownOption := imageCommandOptionType(99)
	result := unknownOption.String()
	expected := "ApplicationCommandOptionType(99)"
	if result != expected {
		t.Errorf("Expected %s, got %s for unknown option type", expected, result)
	}
}

// Helper function to find an option by name
func findOptionByName(options []*discord.ApplicationCommandOption, name string) *discord.ApplicationCommandOption {
	for _, option := range options {
		if option.Name == name {
			return option
		}
	}
	return nil
}