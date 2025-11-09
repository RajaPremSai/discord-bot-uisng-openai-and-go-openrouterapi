package gpt

import (
	"net/http"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	gptPricePerPromptTokenGPT3Dot5Turbo0613     = 0.0000015
	gptPricePerCompletionTokenGPT3Dot5Turbo0613 = 0.000002

	gptPricePerPromptTokenGPT3Dot5Turbo16K0613     = 0.000003
	gptPricePerCompletionTokenGPT3Dot5Turbo16K0613 = 0.000004

	gptPricePerPromptTokenGPT40613     = 0.00003
	gptPricePerCompletionTokenGPT40613 = 0.00006

	gptPricePerPromptTokenGPT432K0613     = 0.00006
	gptPricePerCompletionTokenGPT432K0613 = 0.00012
)

const (
	gptTruncateLimitGPT3Dot5Turbo0301 = 3500
	gptTruncateLimitGPT40314          = 6500
	gptTruncateLimitGPT432K0314       = 30500
)

func shouldHandleMessageType(t discord.MessageType) bool {
	return t == discord.MessageTypeDefault || t == discord.MessageTypeReply
}

type chatGPTResponse struct {
	content string
	usage   openai.Usage
}

func sendChatGPTRequest(client *openai.Client, cacheItem *MessagesCacheData) (*chatGPTResponse, error) {

}

func getUrlData(client *http.Client, url string) (string, error) {

}

func getContentOrURLData(client *http.Client, s string) (content string, err error) {

}

func parseInteractionReply(discordMessage *discord.Message) (prompt string, context string, model string, temperature *float32) {

}

func modelTruncateLimit(model string) *int {

}

func attachUsageInfo(s *discord.Session, m *discord.Message, usage openai.Usage, model string) {

}

func generateCost(usage openai.Usage, model string) string {

}

func adjustMessageTokens(cacheItem *MessagesCacheData) {

}

func isCacheItemWithinTurncateLimit(cacheItem *MessagesCacheData) (ok bool, count int) {

}

func generateThreadTitleBasedOnInitialPrompt(ctx *bot.Context, client *openai.Client, threadID string, messages []openai.ChatCompletionChoice)
