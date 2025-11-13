package commands

import (
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/commands/dalle"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

const imageCommandName = "image"

func ImageCommand(client *openrouter.Client, imageModel string) *bot.Command {
	return &bot.Command{
		Name:                     imageCommandName,
		Description:              "Generate creative images from textual description",
		DMPermission:             false,
		DefaultMemberPermissions: discord.PermissionViewChannel,
		SubCommands: bot.NewRouter([]*bot.Command{
			dalle.Command(client, imageModel),
		}),
	}
}
