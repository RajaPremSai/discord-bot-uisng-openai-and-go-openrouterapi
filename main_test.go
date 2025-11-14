package main

import (
	"os"
	"testing"
)

func createValidConfig() Config {
	return Config{
		Discord: struct {
			Token          string `yaml:"token"`
			Guild          string `yaml:"guild"`
			RemoveCommands bool   `yaml:"removeCommands"`
		}{
			Token: "test-token",
		},
		OpenRouter: struct {
			APIKey           string   `yaml:"apiKey"`
			BaseURL          string   `yaml:"baseURL"`
			SiteURL          string   `yaml:"siteURL"`
			SiteName         string   `yaml:"siteName"`
			CompletionModels []string `yaml:"completionModels"`
			ImageModels      []string `yaml:"imageModels"`
		}{
			APIKey:           "sk-or-v1-test-key",
			BaseURL:          "https://openrouter.ai/api/v1",
			CompletionModels: []string{"openai/gpt-4"},
			ImageModels:      []string{"openai/dall-e-2"},
		},
	}
}

func createConfigWithMissingDiscordToken() Config {
	config := createValidConfig()
	config.Discord.Token = ""
	return config
}

func createConfigWithMissingAPIKey() Config {
	config := createValidConfig()
	config.OpenRouter.APIKey = ""
	return config
}

func createConfigWithInvalidAPIKey() Config {
	config := createValidConfig()
	config.OpenRouter.APIKey = "sk-invalid-key"
	return config
}

func createConfigWithInvalidBaseURL() Config {
	config := createValidConfig()
	config.OpenRouter.BaseURL = "invalid-url"
	return config
}

func createConfigWithInvalidModel() Config {
	config := createValidConfig()
	config.OpenRouter.CompletionModels = []string{"gpt-4"} // Missing provider prefix
	return config
}

func createConfigWithInvalidImageModel() Config {
	config := createValidConfig()
	config.OpenRouter.ImageModels = []string{"dall-e-2"} // Missing provider prefix
	return config
}

func createConfigWithDefaults() Config {
	return Config{
		Discord: struct {
			Token          string `yaml:"token"`
			Guild          string `yaml:"guild"`
			RemoveCommands bool   `yaml:"removeCommands"`
		}{
			Token: "test-token",
		},
		OpenRouter: struct {
			APIKey           string   `yaml:"apiKey"`
			BaseURL          string   `yaml:"baseURL"`
			SiteURL          string   `yaml:"siteURL"`
			SiteName         string   `yaml:"siteName"`
			CompletionModels []string `yaml:"completionModels"`
			ImageModels      []string `yaml:"imageModels"`
		}{
			APIKey: "sk-or-v1-test-key",
			// BaseURL, CompletionModels, and ImageModels will be set to defaults
		},
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: createValidConfig(),
			wantErr: false,
		},
		{
			name:    "missing discord token",
			config:  createConfigWithMissingDiscordToken(),
			wantErr: true,
			errMsg:  "discord token is required",
		},
		{
			name:    "missing openrouter api key",
			config:  createConfigWithMissingAPIKey(),
			wantErr: true,
			errMsg:  "openRouter API key is required",
		},
		{
			name:    "invalid openrouter api key format",
			config:  createConfigWithInvalidAPIKey(),
			wantErr: true,
			errMsg:  "invalid OpenRouter API key format, must start with 'sk-or-v1-'",
		},
		{
			name:    "invalid base url format",
			config:  createConfigWithInvalidBaseURL(),
			wantErr: true,
			errMsg:  "invalid OpenRouter base URL format, must start with http:// or https://",
		},
		{
			name:    "invalid completion model name format",
			config:  createConfigWithInvalidModel(),
			wantErr: true,
			errMsg:  "invalid OpenRouter completion model name 'gpt-4', must include provider prefix (e.g., 'openai/gpt-4')",
		},
		{
			name:    "invalid image model name format",
			config:  createConfigWithInvalidImageModel(),
			wantErr: true,
			errMsg:  "invalid OpenRouter image model name 'dall-e-2', must include provider prefix (e.g., 'openai/dall-e-2')",
		},
		{
			name:    "config with defaults applied",
			config:  createConfigWithDefaults(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Config.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
				// Check that defaults were applied
				if tt.config.OpenRouter.BaseURL == "" {
					if tt.config.OpenRouter.BaseURL != "https://openrouter.ai/api/v1" {
						t.Errorf("Config.Validate() did not set default BaseURL")
					}
				}
				if len(tt.config.OpenRouter.CompletionModels) == 0 {
					if len(tt.config.OpenRouter.CompletionModels) != 1 || tt.config.OpenRouter.CompletionModels[0] != "openai/gpt-3.5-turbo" {
						t.Errorf("Config.Validate() did not set default CompletionModels")
					}
				}
				if len(tt.config.OpenRouter.ImageModels) == 0 {
					if len(tt.config.OpenRouter.ImageModels) != 1 || tt.config.OpenRouter.ImageModels[0] != "openai/dall-e-2" {
						t.Errorf("Config.Validate() did not set default ImageModels")
					}
				}
			}
		})
	}
}

func TestConfig_ReadFromFile(t *testing.T) {
	// Create a temporary test config file
	testConfig := `discord:
  token: "test-token"
  guild: "test-guild"
  removeCommands: true

openRouter:
  apiKey: "sk-or-v1-test-key"
  baseURL: "https://openrouter.ai/api/v1"
  siteURL: "https://test-site.com"
  siteName: "Test Bot"
  completionModels:
    - "openai/gpt-4"
    - "openai/gpt-3.5-turbo"
  imageModels:
    - "openai/dall-e-2"
    - "openai/dall-e-3"
`

	// Write test config to temporary file
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tmpFile.Close()

	// Test reading the config
	config := &Config{}
	err = config.ReadFromFile(tmpFile.Name())
	if err != nil {
		t.Errorf("Config.ReadFromFile() error = %v", err)
		return
	}

	// Verify the config was parsed correctly
	if config.Discord.Token != "test-token" {
		t.Errorf("Expected Discord.Token = 'test-token', got %v", config.Discord.Token)
	}
	if config.Discord.Guild != "test-guild" {
		t.Errorf("Expected Discord.Guild = 'test-guild', got %v", config.Discord.Guild)
	}
	if config.OpenRouter.APIKey != "sk-or-v1-test-key" {
		t.Errorf("Expected OpenRouter.APIKey = 'sk-or-v1-test-key', got %v", config.OpenRouter.APIKey)
	}
	if config.OpenRouter.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("Expected OpenRouter.BaseURL = 'https://openrouter.ai/api/v1', got %v", config.OpenRouter.BaseURL)
	}
	if len(config.OpenRouter.CompletionModels) != 2 {
		t.Errorf("Expected 2 completion models, got %d", len(config.OpenRouter.CompletionModels))
	}
	if len(config.OpenRouter.ImageModels) != 2 {
		t.Errorf("Expected 2 image models, got %d", len(config.OpenRouter.ImageModels))
	}
}

func TestConfig_ReadFromFile_InvalidFile(t *testing.T) {
	config := &Config{}
	err := config.ReadFromFile("nonexistent-file.yaml")
	if err == nil {
		t.Error("Config.ReadFromFile() expected error for nonexistent file but got none")
	}
}

func TestConfig_ReadFromFile_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	invalidYAML := `discord:
  token: "test-token"
  invalid yaml structure
`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	tmpFile.Close()

	config := &Config{}
	err = config.ReadFromFile(tmpFile.Name())
	if err == nil {
		t.Error("Config.ReadFromFile() expected error for invalid YAML but got none")
	}
}