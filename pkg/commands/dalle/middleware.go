package dalle

import (
	"log"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

func imageInteractionResponseMiddleware(ctx *bot.Context) {
	log.Printf("[GID:%s,i.ID:%s] Image interaction invoked by UserID: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, ctx.Interaction.Member.User.ID)

	err := ctx.Respond(&discord.InteractionResponse{
		Type: discord.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("[GID:%s,i.ID:%s] Failed to respond to interation with the error %v", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	ctx.Next()
}
func imageModerationMiddleware(ctx *bot.Context, client *openrouter.Client) {
	log.Printf("[GId : %s,i.ID:%s] Performing interaction moderation middleware\n", ctx.Interaction.GuildID, ctx.Interaction.ID)

	// Note: OpenRouter doesn't have a direct moderation endpoint like OpenAI
	// For now, we'll skip moderation and let OpenRouter handle content filtering
	// TODO: Implement alternative content moderation if needed
	log.Printf("[GID: %s, i.ID:%s] Skipping moderation check - OpenRouter handles content filtering\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
	ctx.Next()
}
