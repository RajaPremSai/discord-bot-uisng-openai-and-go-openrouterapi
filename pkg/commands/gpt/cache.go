package gpt

import (
	"strings"
	
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
)

type IgnoredChannelsCache map[string]struct{}

type MessagesCache struct {
	*lru.Cache[string, *MessagesCacheData]
}

type MessagesCacheData struct {
	Messages      []openrouter.ChatCompletionMessage
	SystemMessage *openrouter.ChatCompletionMessage
	Model         string
	Temperature   *float32
	TokenCount    int
}

// ValidateOpenRouterModel checks if the model name is in valid OpenRouter format
func (c *MessagesCacheData) ValidateOpenRouterModel() bool {
	if c.Model == "" {
		return false
	}
	
	// OpenRouter models typically follow "provider/model" format
	// but also accept direct model names for backward compatibility
	if strings.Contains(c.Model, "/") {
		parts := strings.Split(c.Model, "/")
		return len(parts) == 2 && parts[0] != "" && parts[1] != ""
	}
	
	// Direct model names are also valid (e.g., "gpt-4", "gpt-3.5-turbo")
	return true
}

// GetNormalizedModelName returns a user-friendly model name for display
func (c *MessagesCacheData) GetNormalizedModelName() string {
	return normalizeOpenRouterModelName(c.Model)
}

// GetBaseModelName extracts the base model name for token counting and limits
func (c *MessagesCacheData) GetBaseModelName() string {
	return extractBaseModel(c.Model)
}

func NewMessagesCache(size int) (*MessagesCache, error) {
	lruCache, err := lru.New[string, *MessagesCacheData](size)
	if err != nil {
		return nil, err
	}

	return &MessagesCache{
		Cache: lruCache,
	}, nil
}
