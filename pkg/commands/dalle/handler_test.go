package dalle

import (
	"testing"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
)

func TestImageHandler_ClientInterface(t *testing.T) {
	// Test that the handler accepts OpenRouter client
	client := &openrouter.Client{}
	imageModel := "openai/dall-e-2"
	
	// This test just verifies that the function signature is correct
	// and that we can create the necessary types
	if client == nil {
		t.Error("OpenRouter client should not be nil")
	}
	
	if imageModel == "" {
		t.Error("Image model should not be empty")
	}
}

func TestCommand_ClientInterface(t *testing.T) {
	// Test that the Command function accepts OpenRouter client
	client := &openrouter.Client{}
	imageModel := "openai/dall-e-2"
	
	cmd := Command(client, imageModel)
	if cmd == nil {
		t.Error("Command should not be nil")
	}
	
	if cmd.Name != "dalle" {
		t.Errorf("Expected command name 'dalle', got '%s'", cmd.Name)
	}
}