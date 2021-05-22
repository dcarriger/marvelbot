package server

import (
	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "card",
			Description: "Retrieve a Marvel Champions card or cards",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "image",
					Description: "Fetches the requested card image(s)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "card-names",
							Description: "Card name(s), separated by semi-colons (e.g., Relentless Assault;Follow Through)",
							Required:    true,
						},
						/*
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "private",
							Description: "Make the response private (only visible to you)",
							Required:    false,
						},
						 */
					},
				},
			},
		},
	}
)
