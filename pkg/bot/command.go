// handler of commands
package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

type Handler interface {
	HandlerCommand(ctx *Context)
}

type HandlerFunc func(ctx *Context)

func (f HandlerFunc) HandleCommand(ctx *Context) { f(ctx) }

type MessageHandler interface {
	HandleMessageCommand(ctx *MessageContext)
}

type MessageHandlerFunc func(ctx *MessageContext)

func (f MessageHandlerFunc) HandleMessageCommand(ctx *MessageContext) { f(ctx) }

type Command struct {
	Name                     string
	Description              string
	DMPermission             bool
	DefaultMemberPermissions int64
	Options                  []*discord.ApplicationCommandOption
	Type                     discord.ApplicationCommandType
	Handler                  Handler
	Middlewares              []Handler
	MessageHandler           MessageHandler
	Subcommands              *Router
}
