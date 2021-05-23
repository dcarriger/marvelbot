package server

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"marvelbot/card"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Aspect contains the name and color associated with a Marvel Champions Aspect
type Aspect struct {
	Name  string `json:"name" yaml:"name"`
	Color int    `json:"color" yaml:"color"`
}

// Hero contains the name of the hero and their S3 image URL
type Hero struct {
	Name  string `json:"name" yaml:"name"`
	Image string `json:"image" yaml:"image"`
}

// Villain is the encounter that we will be using
type Villain struct {
	Name               string   `json:"name" yaml:"name"`
	Image              string   `json:"image" yaml:"image"`
	ModuleCount        int      `json:"module_count" yaml:"module_count"`
	RecommendedModules []string `json:"recommended_modules" yaml:"recommended_modules"`
	RequiredModules    []string `json:"required_modules" yaml:"required_modules"`
}

const (
	Aggression = 0x78141b
	Basic      = 0x8c8c8c
	Justice    = 0xa09320
	Leadership = 0x3ea0b2
	Protection = 0x59aa36
)

var (
	Aspects = []*Aspect{
		{
			Name:  "Aggression",
			Color: Aggression,
		},
		{
			Name:  "Justice",
			Color: Justice,
		},
		{
			Name:  "Leadership",
			Color: Leadership,
		},
		{
			Name:  "Protection",
			Color: Protection,
		},
	}
	Heroes = []*Hero{
		{"Ant-Man", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc12en/1A.png"},
		{"Black Panther", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/40A.png"},
		{"Black Widow", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc07en/1A.png"},
		{"Captain America", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc04en/1A.png"},
		{"Captain Marvel", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/10A.png"},
		{"Doctor Strange", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc08en/1A.png"},
		{"Gamora", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc18en/1A.png"},
		{"Groot", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/1B.png"},
		{"Hawkeye", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/1A.png"},
		{"Hulk", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc09en/1A.png"},
		{"Iron Man", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/29A.png"},
		{"Ms. Marvel", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc05en/1A.png"},
		{"Quicksilver", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc14en/1A.png"},
		{"Rocket Raccoon", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/29B.png"},
		{"Scarlet Witch", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc15en/1A.png"},
		{"She-Hulk", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/19A.png"},
		{"Spider-Man", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/1A.png"},
		{"Spider-Woman", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/31A.png"},
		{"Star-Lord", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc17en/1A.png"},
		{"Thor", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc06en/1A.png"},
		{"Wasp", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc13en/1A.png"},
	}
	Villains = []*Villain{
		{"Rhino", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/94.png", 1, []string{"Bomb Scare"}, []string{}},
		{"Klaw", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/113.png", 1, []string{"Masters of Evil"}, []string{}},
		{"Ultron", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc01en/134.png", 1, []string{"Under Attack"}, []string{}},
		{"Green Goblin (Risky Business)", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc02en/1A.png", 1, []string{"Goblin Gimmicks"}, []string{}},
		{"Green Goblin (Mutagen Formula)", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc02en/14.png", 1, []string{"Goblin Gimmicks"}, []string{}},
		{"Wrecking Crew", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc03en/2.png", 0, []string{}, []string{}},
		{"Crossbones", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/58.png", 2, []string{"Hydra Assault", "Weapon Master"}, []string{"Experimental Weapons", "Legions of Hydra"}},
		{"Absorbing Man", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/76.png", 1, []string{"Hydra Patrol"}, []string{}},
		{"Taskmaster", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/93.png", 1, []string{"Weapon Master"}, []string{"Hydra Patrol"}},
		{"Zola", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/109.png", 1, []string{"Under Attack"}, []string{}},
		{"Red Skull", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc10en/125.png", 2, []string{"Hydra Assault", "Hydra Patrol"}, []string{}},
		{"Kang", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc11en/1.png", 1, []string{"Temporal"}, []string{}},
		{"Drang", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/58.png", 1, []string{"Band of Badoon"}, []string{"Ship Command"}},
		{"The Collector (Infiltrate the Museum)", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/70.png", 1, []string{"Menagerie Medley"}, []string{"Galactic Artifacts"}},
		{"The Collector (Escape the Museum)", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/80A.png", 1, []string{"Menagerie Medley"}, []string{"Galactic Artifacts"}},
		{"Nebula", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/88.png", 1, []string{"Space Pirates"}, []string{"Ship Command", "The Power Stone"}},
		{"Ronan", "https://marvel-champions-cards.s3.us-west-2.amazonaws.com/mc16en/103.png", 1, []string{"Kree Militants"}, []string{"Ship Command", "The Power Stone"}},
	}
	Modules = []string{
		"Bomb Scare",
		"Masters of Evil",
		"Under Attack",
		"Legions of Hydra",
		"The Doomsday Chair",
		"Goblin Gimmicks",
		"A Mess of Things",
		"Running Interference",
		"Power Drain",
		"Weapon Master",
		"Hydra Assault",
		"Hydra Patrol",
		"Temporal",
		"Master of Time",
		"Anachronauts",
		"Band of Badoon",
		"Menagerie Medley",
		"Space Pirates",
		"Kree Militants",
	}
)

// CardHandler serves the "card" slash command and subcommands.
func (srv *Server) CardHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO - Support private embeds once Discord does?
	// https://github.com/discord/discord-api-docs/issues/2318
	// We must respond to the user within 3 seconds, and in many cases, querying
	// the card database and putting together a combined image may take longer.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			// Flags: uint64(64),
		},
	})

	if err != nil {
		srv.Logger.Error(fmt.Sprintf("error replying to interaction - %v", err))
		s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Agent <@%s>, HYDRA has corrupted the S.H.I.E.L.D. archives.\nThis issue has been logged for our maintenance team.\n",
				i.Interaction.Member.User.ID),
		})
		return
	}

	// Identify the subcommand and respond accordingly
	switch i.Data.Options[0].Name {
	case "image":
		// Log the request
		srv.Logger.Info(fmt.Sprintf("%s: User %s in Guild %s requested: card image %s", i.ID, i.Interaction.Member.User.Username, i.GuildID, i.Data.Options[0].Options[0].StringValue()))
		commands := strings.SplitN(i.Data.Options[0].Options[0].StringValue(), ";", -1)
		matchedCards := []*card.Card{}
		unmatchedCards := []string{}
		for _, command := range commands {
			// Search for matching cards and append them to the slice
			filter, query := splitCommand(command)
			cards := findCards(filter, query, srv.Cards)
			if len(cards) == 0 {
				unmatchedCards = append(unmatchedCards, query)
				break
			}
			matchedCards = append(matchedCards, cards...)
		}
		// If there are any queries that failed, we need to notify the user.
		// For successful queries, we need to return an attachment (or multiple attachments)
		// for use in a Discord Embed.
		if len(unmatchedCards) > 0 && len(matchedCards) == 0 {
			// TODO - Log what we failed to match
			var content, tense, failedQueries string
			if len(unmatchedCards) > 1 {
				tense = "files"
			} else {
				tense = "file"
			}
			failedQueries = strings.Join(unmatchedCards, "\n")
			content = fmt.Sprintf(
				"Agent <@%s>, the S.H.I.E.L.D database was unable to retrieve the %s you requested:\n\n%s\n\nPlease notify Director <@%s> if you believe this to be an error.",
				i.Interaction.Member.User.ID,
				tense,
				failedQueries,
				Director,
			)
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: content,
			})
			return
		} else if len(unmatchedCards) > 0 && len(matchedCards) > 0 {
			// TODO - implement me
			// TODO - Log what we failed to match
		} else if len(unmatchedCards) == 0 && len(matchedCards) > 0 {
			// We want to return the horizontal and vertical images separately
			horizontalCards := []*card.Card{}
			verticalCards := []*card.Card{}
			for _, c := range matchedCards {
				if c.Horizontal == true {
					horizontalCards = append(horizontalCards, c)
				} else {
					verticalCards = append(verticalCards, c)
				}
			}
			horizontalImage, cardsWithErrors, err := buildImage(horizontalCards)
			verticalImage, cardsWithErrors, err := buildImage(verticalCards)
			// We will return the images to the sender
			if horizontalImage != "" || verticalImage != "" {
				var horizontalFileName, verticalFileName string
				if horizontalImage != "" {
					horizontalFileName = fmt.Sprintf("images/%s", horizontalImage)
				}
				if verticalImage != "" {
					verticalFileName = fmt.Sprintf("images/%s", verticalImage)
				}
				horizontalFile, horizontalErr := os.Open(horizontalFileName)
				defer horizontalFile.Close()
				verticalFile, verticalErr := os.Open(verticalFileName)
				defer verticalFile.Close()
				// FIXME - This has to be a FollowupMessageCreate to allow for attachments. If Discord later allows us to add an attachment to the original message, we'll use that methodology.
				var content, tense string
				if len(matchedCards)-len(cardsWithErrors) > 1 {
					tense = "records"
				} else {
					tense = "record"
				}
				content = fmt.Sprintf(
					"Agent <@%s>, we have located the %s you requested. We are transferring to you now over an encrypted communication channel. This message will self-destruct once you've secured the package.",
					i.Interaction.Member.User.ID,
					tense,
				)
				// Edit the original message to indicate we found the card(s)
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				// Determine which files to return
				files := []*discordgo.File{}
				if horizontalErr == nil {
					x := &discordgo.File{
						Name:        horizontalFileName,
						Reader:      horizontalFile,
						ContentType: "image/jpeg",
					}
					files = append(files, x)
				}
				if verticalErr == nil {
					x := &discordgo.File{
						Name:        verticalFileName,
						Reader:      verticalFile,
						ContentType: "image/jpeg",
					}
					files = append(files, x)
				}
				// Send a message with the attachment
				_, err = s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
					Files: files,
				})
				if err == nil {
					time.Sleep(time.Second * 10)
					err = s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
					if err != nil {
						srv.Logger.Error(fmt.Sprintf("error deleting interaction response - %v", err))
					}
				} else {
					srv.Logger.Error(fmt.Sprintf("error sending attachment - %v", err))
				}
				// We need to delete the temporary file now
				os.Remove(fmt.Sprintf("images/%s", horizontalFile))
				os.Remove(fmt.Sprintf("images/%s", verticalFile))
			}
		}
	case "link":
		// Log the request
		srv.Logger.Info(fmt.Sprintf("%s: User %s in Guild %s requested: card link %s", i.ID, i.Interaction.Member.User.Username, i.GuildID, i.Data.Options[0].Options[0].StringValue()))
		commands := strings.SplitN(i.Data.Options[0].Options[0].StringValue(), ";", -1)
		matchedCards := []*card.Card{}
		unmatchedCards := []string{}
		for _, command := range commands {
			// Search for matching cards and append them to the slice
			filter, query := splitCommand(command)
			cards := findCards(filter, query, srv.Cards)
			if len(cards) == 0 {
				unmatchedCards = append(unmatchedCards, query)
				break
			}
			matchedCards = append(matchedCards, cards...)
		}
		embeds := []*discordgo.MessageEmbed{}
		for _, c := range matchedCards {
			for _, f := range c.Faces {
				var link string
				if f.MarvelCDBURL != nil {
					link = fmt.Sprintf("[%s](%s)", f.Name, *f.MarvelCDBURL)
				} else {
					link = fmt.Sprintf("No MarvelCDB link was found for %s. Either the card does not exist yet in MarvelCDB, or the S.H.I.E.L.D. database is out of date.", f.Name)
				}
				embed := &discordgo.MessageEmbed{
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: *f.ImageURL,
					},
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  f.Name,
							Value: link,
						},
					},
				}
				embeds = append(embeds, embed)
			}
		}
		err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Embeds: embeds,
		})
		if err != nil {
			fmt.Println(err)
		}
	default:
		// Do the thing
	}
}

// MissionHandler serves the "mission" slash command and subcommands.
func (srv *Server) MissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Flags: uint64(64),
		},
	})
	if err != nil {
		srv.Logger.Error(fmt.Sprintf("error replying to interaction - %v", err))
	}
	// Since there are no subcommands, we can jump straight into options
	// These options are all required
	playerCount := i.Data.Options[0].IntValue()
	randomizeHeroes := i.Data.Options[1].BoolValue()
	randomizeAspects := i.Data.Options[2].BoolValue()
	randomizeVillain := i.Data.Options[3].BoolValue()
	randomizeModules := i.Data.Options[4].BoolValue()
	// These options are optional
	var avoidDuplicateAspects = true
	if len(i.Data.Options) >= 6 {
		avoidDuplicateAspects = i.Data.Options[5].BoolValue()
	}
	var modularCount int64 = -1
	if len(i.Data.Options) >= 7 {
		modularCount = i.Data.Options[6].IntValue()
	}

	// Player holds the Hero/Aspect selections for a player
	type Player struct {
		Hero    *Hero     `json:"hero,omitempty"`
		Aspects []*Aspect `json:"aspects,omitempty"`
	}

	// Add the players
	players := []*Player{}
	for player := int64(0); player < playerCount; player++ {
		player := &Player{}
		players = append(players, player)
	}
	// Create the seed
	rand.Seed(time.Now().UnixNano())
	// Determine the Hero for each player
	if randomizeHeroes == true {
		heroes := make([]*Hero, len(Heroes))
		copy(heroes, Heroes)
		for k, _ := range players {
			i := rand.Intn(len(heroes))
			hero := heroes[i]
			players[k].Hero = hero
			heroes = removeHeroIndex(heroes, i)
		}
	}
	// Determine the Aspect for each player
	if randomizeAspects == true {
		aspects := make([]*Aspect, len(Aspects))
		copy(aspects, Aspects)
		for p, v := range players {
			i := rand.Intn(len(aspects))
			aspect := aspects[i]
			players[p].Aspects = append(players[p].Aspects, aspect)
			// Spider-Woman has a second Aspect
			if v.Hero != nil && v.Hero.Name == "Spider-Woman" {
				secondaryAspects := make([]*Aspect, len(Aspects))
				copy(secondaryAspects, Aspects)
				for k, v := range secondaryAspects {
					if aspect == v {
						secondaryAspects = removeAspectIndex(secondaryAspects, k)
					}
				}
				i := rand.Intn(len(secondaryAspects))
				secondAspect := secondaryAspects[i]
				players[p].Aspects = append(players[p].Aspects, secondAspect)
			}
			if avoidDuplicateAspects == true {
				aspects = removeAspectIndex(aspects, i)
			}
		}
	}
	// Determine the villain
	var villain *Villain
	if randomizeVillain == true {
		villains := make([]*Villain, len(Villains))
		copy(villains, Villains)
		i := rand.Intn(len(villains))
		villain = villains[i]
	}
	// Determine the modular encounter sets
	encounterModules := []string{}
	if randomizeModules == true {
		modules := make([]string, len(Modules))
		copy(modules, Modules)
		// How many modules do we need?
		if modularCount <= -1 {
			if villain != nil {
				modularCount = int64(len(villain.RecommendedModules))
			} else {
				modularCount = int64(1)
			}
		}
		if modularCount > int64(len(modules)) {
			modularCount = int64(len(modules))
		}
		// Add random modules to our slice
		for count := int64(0); count < modularCount; count++ {
			i := rand.Intn(len(modules))
			module := modules[i]
			encounterModules = append(encounterModules, module)
			modules = removeStringIndex(modules, i)
		}
	}
	if randomizeModules == false && villain != nil {
		encounterModules = villain.RecommendedModules
	}
	// Return the mission to the players
	embeds := []*discordgo.MessageEmbed{}
	// The villain module comes first
	if villain != nil {
		fields := []*discordgo.MessageEmbedField{}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Villain",
			Value: villain.Name,
		})
		if len(encounterModules) > 0 {
			moduleNames := strings.Join(encounterModules, ", ")
			var fieldName string
			if randomizeModules == true {
				fieldName = "Encounter Modules"
			} else {
				fieldName = "Recommended Encounter Modules"
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  fieldName,
				Value: moduleNames,
			})
		}
		if len(villain.RequiredModules) > 0 {
			moduleNames := strings.Join(villain.RequiredModules, ", ")
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Required Modules",
				Value: moduleNames,
			})
		}
		embed := &discordgo.MessageEmbed{
			Title:       "The Mission",
			Description: fmt.Sprintf("Deputy Director Maria Hill has contacted you with an urgent mission. S.H.I.E.L.D. intelligence has identified an impending threat from %s that requires an immediate response. You have been tasked with neutralizing the threat and minimizing civilian casualties.", villain.Name),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: villain.Image,
			},
			Fields: fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Known issues:\n- Thumbnails don't always load\n- Encounter and Required modules sometimes overlap\n- Missing support for filtering out unwanted Villains/Encounter Modules",
			},
		}
		embeds = append(embeds, embed)
	}
	for k, v := range players {
		fields := []*discordgo.MessageEmbedField{}
		// Select the embed thumbnail
		var thumbnail string
		if v.Hero != nil {
			thumbnail = v.Hero.Image
			field := &discordgo.MessageEmbedField{
				Name:  "Hero",
				Value: v.Hero.Name,
			}
			fields = append(fields, field)
		}
		// Select the embed color
		var color int = Basic
		if len(v.Aspects) > 0 {
			color = v.Aspects[0].Color
		}
		if len(v.Aspects) == 1 {
			field := &discordgo.MessageEmbedField{
				Name:  "Aspect",
				Value: v.Aspects[0].Name,
			}
			fields = append(fields, field)
		} else if len(v.Aspects) == 2 {
			field := &discordgo.MessageEmbedField{
				Name:  "Aspects",
				Value: fmt.Sprintf("%s/%s", v.Aspects[0].Name, v.Aspects[1].Name),
			}
			fields = append(fields, field)
		}
		embed := &discordgo.MessageEmbed{
			Title:     fmt.Sprintf("Player %d", k+1),
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: thumbnail},
			Color:     color,
			Fields:    fields,
		}
		embeds = append(embeds, embed)
	}
	err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Embeds: embeds,
	})
	if err != nil {
		srv.Logger.Error(fmt.Sprintf("error editing mission: %v", err))
	}
}
