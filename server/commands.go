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
					},
				},
				{
					Name:        "link",
					Description: "Returns a MarvelCDB link to the card",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "card-names",
							Description: "Card name(s), separated by semi-colons (e.g., Relentless Assault;Follow Through)",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "mission",
			Description: "S.H.I.E.L.D. is ready to brief you on your next mission, Agent",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "player-count",
					Description: "The number of agents who will be undertaking the mission",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "1p",
							Value: 1,
						},
						{
							Name:  "2p",
							Value: 2,
						},
						{
							Name:  "3p",
							Value: 3,
						},
						{
							Name:  "4p",
							Value: 4,
						},
					},
					Required: true,
				},
				{
					Name:        "randomize-heroes",
					Description: "Let S.H.I.E.L.D. determine which Heroes to use",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
				{
					Name:        "randomize-aspects",
					Description: "Let S.H.I.E.L.D. determine which Aspects to use",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
				{
					Name:        "randomize-encounter",
					Description: "Let S.H.I.E.L.D. determine which Villain to use",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
				{
					Name:        "randomize-modular-sets",
					Description: "Let S.H.I.E.L.D. determine which Modular Encounter Sets to use",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
				{
					Name:        "avoid-duplicate-aspects",
					Description: "Prevent multiple players from receiving the same Aspect",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    false,
				},
				{
					Name:        "modular-encounter-count",
					Description: "Specify the number of modular encounter sets to include",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
	}
)
