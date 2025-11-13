package dalle

import (
	"context"
	"fmt"
	"log"

	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/bot"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/constants"
	"github.com/RajaPremSai/go-openai-dicord-bot/pkg/openrouter"
	discord "github.com/bwmarrin/discordgo"
)

func imageHandler(ctx *bot.Context, client *openrouter.Client, imageModel string) {
	var prompt string
	if option, ok := ctx.Options[imageCommandOptionPrompt.String()]; ok {
		prompt = option.StringValue()
	} else {
		log.Printf("[GID:%s,i.ID:%s] Failed to parse prompt option\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "X Error",
					Description: "Failed to parse prompt option",
					Color:       0xff0000,
				},
			},
		})
		return
	}
	size := imageDefaultSize
	if option, ok := ctx.Options[imageCommandOptionSize.String()]; ok {
		size = option.StringValue()
		log.Printf("[GID : %s,i.ID:%s]Image size provided :%s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, size)
	}

	number := 1
	if option, ok := ctx.Options[imageCommandOptionNumber.String()]; ok {
		number = int(option.IntValue())
		log.Printf("[GID:%s,i.ID:%s] Image number provided :%d\n", ctx.Interaction.GuildID, ctx.Interaction.ID, number)
	}
	log.Printf("[GID:%s,CHID:%s] Dalle request [size:%s,Number:%d]invoked", ctx.Interaction.GuildID, ctx.Interaction.ChannelID, size, number)
	resp, err := client.CreateImage(
		context.Background(),
		openrouter.ImageRequest{
			Prompt:         prompt,
			Model:          imageModel,
			N:              number,
			Size:           size,
			ResponseFormat: "url",
			User:           ctx.Interaction.Member.User.ID,
		},
	)
	if err != nil {
		log.Printf("[GID:%s,i.ID:%s] OpenRouter request CreateImage failed with the error:%v", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ OpenRouter API Failed",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}
	log.Printf("[GID: %s,i.ID:%s] Dalle Reuqest [Size:%s,Number:%d] responded with a data array size %d \n", ctx.Interaction.GuildID, ctx.Interaction.ID, size, number, len(resp.Data))
	var embeds = []*discord.MessageEmbed{
		{
			URL: constants.OpenAIBlackIconURL,
			Author: &discord.MessageEmbedAuthor{
				Name:         prompt,
				IconURL:      ctx.Interaction.Member.AvatarURL("32"),
				ProxyIconURL: constants.OpenAIBlackIconURL,
			},
			Footer: imageCreationUsageEmbedFooter(size, number),
		},
	}

	var buttonComponents []discord.MessageComponent
	for i, data := range resp.Data {
		embeds = append(embeds, &discord.MessageEmbed{
			URL: constants.OpenAIBlackIconURL,
			Image: &discord.MessageEmbedImage{
				URL:    data.URL,
				Width:  256,
				Height: 256,
			},
		})
		buttonComponents = append(buttonComponents, &discord.Button{
			Label: fmt.Sprintf("Image %d", (i + 1)),
			Style: discord.LinkButton,
			URL:   data.URL,
		})
	}
	_, err = ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
		Embeds:     embeds,
		Components: []discord.MessageComponent{discord.ActionsRow{Components: buttonComponents}},
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to send a follow up message with images with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Discord API Error",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}
	// err is nil here (the error branch returned), so just continue with the followup.
	log.Printf("[GID: %s, i.ID: %s] Discord API succeeded\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
	ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
		Content: fmt.Sprintf("> %s", prompt),
		Embeds: []*discord.MessageEmbed{
			{
				Title:       "✅ Discord API Success",
				Description: "Your action completed successfully.",
				Color:       0x00ff00,
			},
		},
	})
}
