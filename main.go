package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/commands"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/commands/gpt"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/constants"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"

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
		ImageModels      []string `yaml:"imageModels"`
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

	// Set default image models if not provided
	if len(c.OpenRouter.ImageModels) == 0 {
		c.OpenRouter.ImageModels = []string{"openai/dall-e-2"}
	}

	// Validate completion model names (should contain provider prefix)
	for _, model := range c.OpenRouter.CompletionModels {
		if !strings.Contains(model, "/") {
			return fmt.Errorf("invalid OpenRouter completion model name '%s', must include provider prefix (e.g., 'openai/gpt-4')", model)
		}
	}

	// Validate image model names (should contain provider prefix)
	for _, model := range c.OpenRouter.ImageModels {
		if !strings.Contains(model, "/") {
			return fmt.Errorf("invalid OpenRouter image model name '%s', must include provider prefix (e.g., 'openai/dall-e-2')", model)
		}
	}

	return nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var (
	discordBot      *bot.Bot
	openrouterClient *openrouter.Client

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
		log.Printf("Initializing OpenRouter client with base URL: %s", config.OpenRouter.BaseURL)
		
		openrouterClient = openrouter.NewClientWithConfig(openrouter.ClientConfig{
			APIKey:   config.OpenRouter.APIKey,
			BaseURL:  config.OpenRouter.BaseURL,
			SiteURL:  config.OpenRouter.SiteURL,
			SiteName: config.OpenRouter.SiteName,
		})
		
		log.Printf("OpenRouter client initialized successfully")
		if config.OpenRouter.SiteURL != "" {
			log.Printf("OpenRouter site URL configured: %s", config.OpenRouter.SiteURL)
		}
		if config.OpenRouter.SiteName != "" {
			log.Printf("OpenRouter site name configured: %s", config.OpenRouter.SiteName)
		}
		
		// Test OpenRouter client connection
		log.Printf("Testing OpenRouter API connection...")
		ctx := context.Background()
		if err := openrouterClient.Ping(ctx); err != nil {
			log.Printf("Warning: OpenRouter API connection test failed: %v", err)
			log.Printf("Continuing with initialization, but API calls may fail")
		} else {
			log.Printf("OpenRouter API connection test successful")
		}
		
		// Log available models
		log.Printf("Configured completion models: %v", config.OpenRouter.CompletionModels)
		log.Printf("Configured image models: %v", config.OpenRouter.ImageModels)
		
		// Get default image model (first one in the list)
		defaultImageModel := config.OpenRouter.ImageModels[0]
		log.Printf("Using default image model: %s", defaultImageModel)
		
		// Register commands with OpenRouter client
		log.Printf("Registering chat command with OpenRouter client")
		discordBot.Router.Register(commands.ChatCommand(&commands.ChatCommandParams{
			OpenRouterClient:     openrouterClient,
			CompletionModels:     config.OpenRouter.CompletionModels,
			GPTMessagesCache:     gptMessagesCache,
			IgnoredChannelsCache: &ignoredChannelsCache,
		}))
		
		log.Printf("Registering image command with OpenRouter client")
		discordBot.Router.Register(commands.ImageCommand(openrouterClient, defaultImageModel))
		
		log.Printf("OpenRouter client initialization and command registration completed")
	} else {
		log.Printf("Warning: OpenRouter API key not configured, AI commands will not be available")
	}
	log.Printf("Loaded Discord Token: %s", config.Discord.Token)
	discordBot.Router.Register(commands.InfoCommand())
	discordBot.Run(config.Discord.Guild, config.Discord.RemoveCommands)
}