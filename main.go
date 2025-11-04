package main

import (
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert/yaml"
)

type Config struct{
	Discord struct{
		Token string `yaml:"token"`
		Guild string `yaml:"guild"`
		RemoveCommands bool `yaml:"removeCommands"`
	}`yaml:"discord"`
	OpenAI struct{
		APIKey string `yaml:"apiKey"`
		CompletionModels []string `yaml:"completionModels"`
	}`yaml:"openAI"`
}

func (c *Config) ReadFromFile(file string)error{
	data,err:=os.ReadFile(file)
	if err!=nil{
		return err
	}
	err=yaml.Unmarshal(data,c)
	if err!=nil{
		return err
	}
	return err
}

func init(){
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var(
	discordBot *bot.Bot
	openaiClient *openai.Client

	gptMessageCache *gpt.MessagesCache
	ignoredChannelCache=make(gpt.IgnoredChannelsCache)
)

func main(){
	config :=&Config{}
	err:=config.ReadFromFile("credentials.yaml")
	if err!=nil{
		log.Fatalf("Error reading credentials.yaml: %v",err)
	}
	gptMessageCache,err = gpt.NewMessagesCache(constants.DiscordThreadsCacheSize)
	if err!=nil{
		log.Fatalf("Error initializing GPTMessageCache: %v",err)
	}
	discordBot,err := bot.NewBot(config.Discord.Token)
	if err!=nil{
		log.Fatalf("Inavalid parameters:%v",err)
	}
	if config.OpenAI.APIKey !=""{
		openaiClient=openai.NewClient(config.OpenAI.APIKey)
		discordBot.Router.Register(commands.ChatCommand(&commands.ChatCommandParams{OpenAIClient:           openaiClient,
			OpenAICompletionModels: config.OpenAI.CompletionModels,
			GPTMessagesCache:       gptMessagesCache,
			IgnoredChannelsCache:   &ignoredChannelsCache,}))
		discordBot.Router.Register(commands.ImageCommand(openaiClient))
	}
	discordBot.Router.Register(commands.InfoCommand())
	discordBot.Run(config.Discord.Guild,config.Discord.RemoveCommands)
}