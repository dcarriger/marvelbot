package server

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"marvelbot/card"
	"os"
	"strings"
	"time"
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
		// TODO - Log error here
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
			if err != nil {
				// TODO - Handle error
			}
			verticalImage, cardsWithErrors, err := buildImage(verticalCards)
			// We will return the images to the sender
			if horizontalImage != "" || verticalImage != "" {
				// Open the file so we can attach it in the response
				horizontalFile, err := os.Open("images/" + horizontalImage)
				if err != nil {
					// TODO - Handle error
				}
				defer horizontalFile.Close()
				verticalFile, err := os.Open("images/" + verticalImage)
				if err != nil {
					// TODO - Handle error
				}
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
				if horizontalFile != nil {
					x := &discordgo.File{
						Name:        fmt.Sprintf("images/%s", horizontalImage),
						Reader:      horizontalFile,
						ContentType: "image/jpeg",
					}
					files = append(files, x)
				}
				if verticalFile != nil {
					x := &discordgo.File{
						Name:        fmt.Sprintf("images/%s", verticalImage),
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
						// TODO - Handle the error
					}
				} else {
					// TODO - Handle the error
				}
				// We need to delete the temporary file now
				os.Remove(fmt.Sprintf("images/%s", horizontalFile))
				os.Remove(fmt.Sprintf("images/%s", verticalFile))
			}
		}
	default:
		// Do the thing
	}
}
