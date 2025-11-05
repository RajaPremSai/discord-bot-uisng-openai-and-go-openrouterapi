//logic for registering the bot,handling messages, handling interactions

package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

type Router struct {
	commands           map[string]*Command
	registeredCommands []*discord.ApplicationCommand
}

func NewRouter(initial []*Command) (r *Router) {
	r = &Router{commands: make(map[string]*Command, len(inital))}
	for _, cmd := range initial {
		r.Register(cmd)
	}

	return
}

func (r *Router) Register(cmd *Command) {
	if _, ok := r.commands[cmd.Name]; !ok {
		r.commands[cmd.Name] = cmd
	}
}

func(r *Router)Get(name string)*Command{

}

func(r *Router)List()(list []*Command){

}

func(r *Router)Count()(c int){
	
}

func( r* Router)getSubcommand(cmd *Command,opt *discord.ApplicationCommandInteractionDataOption,parent []Handler)(){

}

func ( r *Router)getMessageHandlers(cmd *Command)[]MessageHandler{

}

func (r *Router) HandleInteraction(s *discord.Session, i *discord.InteractionCreate) {

}

func (r *Router) HandleMessage(s *discord.Session, m *discord.MessageCreate) {

}

func (r *Router) Sync(s *discord.Session, guild string)(err error){

}

func(r *Router)ClearCommands(s *discord.Session,guild string)(Errors []error){

}