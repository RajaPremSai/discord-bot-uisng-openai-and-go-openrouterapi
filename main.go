package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/commands"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/commands/gpt"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/constants"
	"github.com/sashabaranov/go-openai"

	// "github.com/stretchr/testify/assert/yaml"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Discord struct {
		Token          string `yaml:"token"`
		Guild          string `yaml:"guild"`
		RemoveCommands bool   `yaml:"removeCommands"`
	} `yaml:"discord"`
	OpenRouter struct {
		APIKey           string   `yaml:"apiKey"`
		BaseURL          string   `yaml:"baseURL"`
		SiteURL          string   `yaml:"siteURL"`
		SiteName         string   `yaml:"siteName"`
		CompletionModels []string `yaml:"completionModels"`
	} `yaml:"openRouter"`
}

func (c *Config) ReadFromFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return c.Validate()
}

func (c *Config) Validate() error {
	// Validate Discord configuration
	if c.Discord.Token == "" {
		return fmt.Errorf("discord token is required")
	}

	// Validate OpenRouter configuration
	if c.OpenRouter.APIKey == "" {
		return fmt.Errorf("openRouter API key is required")
	}

	// Validate API key format (OpenRouter keys start with "sk-or-v1-")
	if !strings.HasPrefix(c.OpenRouter.APIKey, "sk-or-v1-") {
		return fmt.Errorf("invalid OpenRouter API key format, must start with 'sk-or-v1-'")
	}

	// Set default base URL if not provided
	if c.OpenRouter.BaseURL == "" {
		c.OpenRouter.BaseURL = "https://openrouter.ai/api/v1"
	}

	// Validate base URL format
	if !strings.HasPrefix(c.OpenRouter.BaseURL, "http://") && !strings.HasPrefix(c.OpenRouter.BaseURL, "https://") {
		return fmt.Errorf("invalid OpenRouter base URL format, must start with http:// or https://")
	}

	// Set default completion models if not provided
	if len(c.OpenRouter.CompletionModels) == 0 {
		c.OpenRouter.CompletionModels = []string{"openai/gpt-3.5-turbo"}
	}

	// Validate model names (should contain provider prefix)
	for _, model := range c.OpenRouter.CompletionModels {
		if !strings.Contains(model, "/") {
			return fmt.Errorf("invalid OpenRouter model name '%s', must include provider prefix (e.g., 'openai/gpt-4')", model)
		}
	}

	return nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var (
	discordBot   *bot.Bot
	openaiClient *openai.Client

	gptMessagesCache     *gpt.MessagesCache
	ignoredChannelsCache = make(gpt.IgnoredChannelsCache)
)

func main() {
	config := &Config{}
	err := config.ReadFromFile("credentials.yaml")
	if err != nil {
		log.Fatalf("Error reading credentials.yaml: %v", err)
	}
	gptMessagesCache, err = gpt.NewMessagesCache(constants.DiscordThreadsCacheSize)
	if err != nil {
		log.Fatalf("Error initializing GPTMessageCache: %v", err)
	}
	discordBot, err := bot.NewBot(config.Discord.Token)
	if err != nil {
		log.Fatalf("Inavalid parameters:%v", err)
	}
	if config.OpenRouter.APIKey != "" {
		openaiClient = openai.NewClient(config.OpenRouter.APIKey)
		discordBot.Router.Register(commands.ChatCommand(&commands.ChatCommandParams{OpenAIClient: openaiClient,
			OpenAICompletionModels: config.OpenRouter.CompletionModels,
			GPTMessagesCache:       gptMessagesCache,
			IgnoredChannelsCache:   &ignoredChannelsCache}))
		discordBot.Router.Register(commands.ImageCommand(openaiClient))
	}
	log.Printf("Loaded Discord Token: %s", config.Discord.Token)
	discordBot.Router.Register(commands.InfoCommand())
	discordBot.Run(config.Discord.Guild, config.Discord.RemoveCommands)
}