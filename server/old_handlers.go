package server

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	gim "github.com/ozankasikci/go-image-merge"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"image/png"
	"marvelbot/card"
	"marvelbot/rule"
	"math"
	"os"
	"regexp"
	"strings"
)

// These constants map to the color codes used by different Aspects.
const (
	IMAGE_BASEDIR = "images"
)

// This function will be called every time a new message is created on any channel that the authenticated bot has
// access to (due to the DiscordGo AddHandler).
func (srv *Server) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	// Help handler
	if m.Content == "[[!help]]" {
		srv.HandleHelp(s, m, m.Author)
	}

	// Command handler
	cardRegexp := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	if cardRegexp.MatchString(m.Content) {
		srv.HandleCommands(s, m, m.Author)
	}
}

// HandleHelp is a function that returns supported bot commands.
func (srv *Server) HandleHelp(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) {
	// Configure logger
	logError := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
		"user":   u.Username,
		"level":  "error",
	})

	// Build our embed fields for the response
	fields := []*discordgo.MessageEmbedField{}

	// Append all bot commands
	fetchCard := &discordgo.MessageEmbedField{
		Name: "[[<card name>]]",
		Value: "Fetches <card name> from MarvelCDB.com and displays the image (if available) or links to the card, " +
			"e.g. [[Lockjaw]]. Multiple cards can be requested in a single message, e.g. [[Peanut Butter]] [[Jelly]]. " +
			"Some filters are also supported in the format of [[<filter>:<search term>]], such as " +
			"[[set:A Mess of Things]].\n\nCurrently supported filters: set",
		Inline: false,
	}

	displayHelp := &discordgo.MessageEmbedField{
		Name:   "!help",
		Value:  "Displays this help message.",
		Inline: false,
	}

	fields = append(fields, fetchCard, displayHelp)

	ms := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       s.State.User.Username,
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       Basic,
			Description: "This bot provides useful resources for Marvel Champions: The Card Game.\n\nUsage:",
			Fields:      fields,
		},
	}

	_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
	if err != nil {
		logError.Errorf("error sending message: %v", err)
	}
	return
}

// HandleCommands is a function that handles requests for Cards, Rules, and other content.
func (srv *Server) HandleCommands(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) {
	// Configure logger for failed Discord message sends
	logError := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
		"user":   u.Username,
		"level":  "error",
	})
	// TODO - remove this once it's in use
	// TODO - Or we are using a better logging package
	_ = logError

	// We need to match all bot commands that were invoked and send them to the appropriate handlers
	cardRegexp := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	results := cardRegexp.FindAllString(m.Content, -1)
	// matchedCards and matchedRules hold those respective objects
	matchedCards := []*card.Card{}
	matchedInfo := []*card.Card{}
	matchedRules := []*rule.Rule{}
	// unmatchedCommands holds anything the bot was unable to find
	unmatchedCommands := []string{}

	// We are going to iterate over the command results and identify what to do with them
	for _, command := range results {
		// The command brackets are no longer needed
		command = trimCommand(command)
		// Query is the content the user is searching for, e.g., Heimdall or Lockjaw
		// Examples: [[Heimdall]], [[Lockjaw]]
		var query string
		// Filter is narrowing the user's query results down, e.g., Ally or Rule
		// Examples: [[Ally:Spider-Man]], [[Rule:Villain Phase]]
		var filter string
		// By default, anything without a filter is considered to be a search for a card
		// TODO - Support multiple filters? e.g., type=Ally, set=Galaxy's Most Wanted
		if strings.Contains(command, ":") {
			commandParts := strings.Split(command, ":")
			filter = strings.ToLower(commandParts[0])
			query = strings.ToLower(commandParts[1])
		} else {
			query = strings.ToLower(command)
		}

		// If the query was too short, reject it
		if len(query) < 3 {
			msg := fmt.Sprintf(
				"Director %s, the S.H.I.E.L.D. database requires that your query contain 3 or more characters:\n%s",
				u.Mention(), query)
			_, err := s.ChannelMessageSend(m.ChannelID, msg)
			if err != nil {
				logError.Errorf("error sending message: %v", err)
			}
			return
		}

		// Based on the filter, we'll handle the command differently
		switch filter {
		case "hb", "homebrew":
			cards := findCards(filter, query, srv.Homebrew)
			// We didn't find a card, so we'll add the command to the list of failed commands
			if len(cards) == 0 {
				unmatchedCommands = append(unmatchedCommands, query)
				break
			}
			// We found a card (or cards), so we'll add them to the list of Cards to return to the user
			matchedCards = append(matchedCards, cards...)
		case "info":
			cards := findCards(filter, query, srv.Cards)
			// We didn't find a card, so we'll add the command to the list of failed commands
			if len(cards) == 0 {
				unmatchedCommands = append(unmatchedCommands, query)
				break
			}
			// We found a card (or cards), so we'll add them to the list of Cards to return to the user
			matchedInfo = append(matchedInfo, cards...)
		case "rule", "rules":
			r := findRule(query, srv.Rules)
			// We didn't find a rule, so we'll add the command to the list of failed commands
			if r == nil {
				unmatchedCommands = append(unmatchedCommands, query)
				break
			}
			// We found a rule, so we'll add it to the list of Rules to return to the user
			matchedRules = append(matchedRules, r)
		case "set":
			cards := findCards(filter, query, srv.Cards)
			// We didn't find a card, so we'll add the command to the list of failed commands
			if len(cards) == 0 {
				unmatchedCommands = append(unmatchedCommands, query)
				break
			}
			// We found a card (or cards), so we'll add them to the list of Cards to return to the user
			matchedCards = append(matchedCards, cards...)
		default:
			cards := findCards(filter, query, srv.Cards)
			// We didn't find a card, so we'll add the command to the list of failed commands
			if len(cards) == 0 {
				unmatchedCommands = append(unmatchedCommands, query)
				break
			}
			// We found a card (or cards), so we'll add them to the list of Cards to return to the user
			matchedCards = append(matchedCards, cards...)
		}
	}

	// Now, we need to start returning these objects to the caller
	if len(unmatchedCommands) > 0 {
		msg := fmt.Sprintf(
			"Director %s, the S.H.I.E.L.D. database has no records for the following queries:\n%s",
			u.Mention(), strings.Join(unmatchedCommands, "\n"))
		_, err := s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			logError.Errorf("error sending message: %v", err)
		}
	}

	// Return all matched rules
	if len(matchedRules) > 0 {
		sendRulesMessages(srv, s, m, u, matchedRules)
	}

	// Return all matched "info" requests
	if len(matchedInfo) > 0 {
		sendCardInfoMessage(srv, s, m, u, matchedInfo)
	}

	// Return all matched cards
	if len(matchedCards) > 0 {
		sendCardMessages(srv, s, m, u, matchedCards)
	}
}

// sendCardInfoMessage will send metadata about a group of cards to the channel
func sendCardInfoMessage(srv *Server, s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, cards []*card.Card) {
	// Configure logger for failed Discord message sends
	logError := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
		"user":   u.Username,
		"level":  "error",
	})

	// Create a field for each card
	// TODO - I think we are limited on how many fields we can have
	// This might be a problem in the future
	var fields []*discordgo.MessageEmbedField

	// Add each card as a field
	for _, c := range cards {
		for _, face := range c.Faces {
			field := &discordgo.MessageEmbedField{}
			field.Name = face.Name
			if face.MarvelCDBURL != nil {
				field.Value = fmt.Sprintf("**[Click here for image](%s)**\n", *face.MarvelCDBURL)
			} else if face.ImageURL != nil {
				field.Value = fmt.Sprintf("**[Click here for image](%s)**\n", *face.ImageURL)
			} else {
				field.Value = "Image(s): No images available\n"
			}
			field.Value += fmt.Sprintf("Type: %s\n", face.Type)
			if len(face.Aspect) > 0 {
				field.Value += fmt.Sprintf("Aspect(s): %s\n", strings.Join(face.Aspect, ","))
			}
			if len(c.Packs) > 0 {
				field.Value += "Packs:\n"
				for _, pack := range c.Packs {
					field.Value += fmt.Sprintf("• %s - %s (x%d)\n", pack.SKU, pack.Name, *pack.Quantity)
				}
			}
			fields = append(fields, field)
		}
	}

	// Send a message back to the channel
	ms := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:  0x78141b,
			Fields: fields,
		},
	}

	_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
	if err != nil {
		logError.Errorf("error sending message: %v", err)
	}
}

// sendCardMessage will send an embedded Cards object to the channel
func sendCardMessages(srv *Server, s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, cards []*card.Card) {
	// Configure logger for failed Discord message sends
	logError := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
		"user":   u.Username,
		"level":  "error",
	})

	// We want to build different grids for horizontal and vertical images so we need to separate them
	// https://github.com/ozankasikci/go-image-merge#examples
	verticalGrids := []*gim.Grid{}
	horizontalGrids := []*gim.Grid{}

	// If we have any cards missing images, we'll use this to return a link to the card instead
	cardsWithErrors := []*card.Card{}

	// Generate the grids
	for _, c := range cards {
		err := c.DownloadImages()
		if err != nil {
			logError.Errorf("error downloading %v card image: %v", c.Names, err)
			cardsWithErrors = append(cardsWithErrors, c)
			continue
		}
		for _, cardFace := range c.Faces {
			// Copied from card.DownloadImages
			imageSlice := strings.Split(*cardFace.ImageURL, "/")
			imageName := imageSlice[len(imageSlice)-1]
			imageDir := fmt.Sprintf("%s/%s", IMAGE_BASEDIR, c.Packs[0].SKU)
			imagePath := fmt.Sprintf("%s/%s", imageDir, imageName)

			// Add cards to their respective Grid based on orientation (vertical or horizontal)
			grid := &gim.Grid{
				ImageFilePath: imagePath,
			}
			if c.Horizontal == true {
				horizontalGrids = append(horizontalGrids, grid)
			} else {
				verticalGrids = append(verticalGrids, grid)
			}
		}
	}

	// If we have a 0-length Grid, all images have failed
	if len(horizontalGrids) == 0 && len(verticalGrids) == 0 {
		var names = []string{}
		for _, card := range cardsWithErrors {
			for _, face := range card.Faces {
				names = append(names, face.Name)
			}
		}
		msg := fmt.Sprintf(
			"Director %s, the S.H.I.E.L.D. database was unable to return images for the following:\n%s",
			u.Mention(), strings.Join(names, "\n"))
		_, err := s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			logError.Errorf("error sending message: %w", err)
		}
		return
	}

	// Merge the images into a grid
	// TODO - Make this a function that returns a filename?
	// We'll start with vertical images
	var verticalFileName string
	var verticalFile *os.File
	if len(verticalGrids) > 0 {
		// # of images per row
		var horizontalCount, verticalCount int
		if len(verticalGrids) > 3 {
			horizontalCount = 3
			verticalCount = int(math.Ceil(float64(len(verticalGrids)) / 3.0))
		} else {
			horizontalCount = len(verticalGrids)
			verticalCount = 1
		}

		rgba, err := gim.New(verticalGrids, horizontalCount, verticalCount).Merge()
		if err != nil {
			logError.Errorf("error merging vertical images: %v", err)
		}

		// Create a guid to use for our new image
		guid := ksuid.New()
		verticalFileName = fmt.Sprintf("temp_%s.png", guid.String())

		// Save the output to PNG
		verticalFile, err = os.Create(fmt.Sprintf("%s/%s", IMAGE_BASEDIR, verticalFileName))
		if err != nil {
			logError.Errorf("error saving file: %v", err)
		}
		err = png.Encode(verticalFile, rgba)
		if err != nil {
			logError.Errorf("error encoding PNG: %v", err)
		}
		defer verticalFile.Close()

		// Rewind to beginning of file after writing to it
		_, err = verticalFile.Seek(0, 0)
		if err != nil {
			msg := fmt.Sprintf("error rewinding file: %v", err)
			_, err = s.ChannelMessageSend(m.ChannelID, msg)
			if err != nil {
				logError.Errorf("error sending message: %v", err)
			}
			return
		}
	}

	// Now, the horizontal images
	// These will be 2 per line
	var horizontalFileName string
	var horizontalFile *os.File
	if len(horizontalGrids) > 0 {
		// # of images per row
		var horizontalCount, verticalCount int
		if len(horizontalGrids) > 2 {
			horizontalCount = 2
			verticalCount = int(math.Ceil(float64(len(horizontalGrids)) / 2.0))
		} else {
			horizontalCount = len(horizontalGrids)
			verticalCount = 1
		}

		rgba, err := gim.New(horizontalGrids, horizontalCount, verticalCount).Merge()
		if err != nil {
			logError.Errorf("error merging horizontal images: %v", err)
		}

		// Create a guid to use for our new image
		guid := ksuid.New()
		horizontalFileName = fmt.Sprintf("temp_%s.png", guid.String())

		// Save the output to PNG
		horizontalFile, err = os.Create(fmt.Sprintf("%s/%s", IMAGE_BASEDIR, horizontalFileName))
		if err != nil {
			logError.Errorf("error saving file: %v", err)
		}
		err = png.Encode(horizontalFile, rgba)
		if err != nil {
			logError.Errorf("error encoding PNG: %v", err)
		}
		defer horizontalFile.Close()

		// Rewind to beginning of file after writing to it
		_, err = horizontalFile.Seek(0, 0)
		if err != nil {
			msg := fmt.Sprintf("error rewinding file: %v", err)
			_, err = s.ChannelMessageSend(m.ChannelID, msg)
			if err != nil {
				logError.Errorf("error sending message: %v", err)
			}
			return
		}
	}

	// Did we have any errors? If so, we'll return them as fields
	var description string
	var fields []*discordgo.MessageEmbedField

	// TODO - reimplement this
	if len(cardsWithErrors) != 0 {
		description = "Some card images could not be returned - these are linked below."
		// Get all the card links
		var cardLinks []string
		// Build the return string
		for _, c := range cardsWithErrors {
			for _, face := range c.Faces {
				if face.MarvelCDBURL != nil {
					link := fmt.Sprintf("[%s](%s)", face.Name, *face.MarvelCDBURL)
					cardLinks = append(cardLinks, link)
				}
			}
		}
		if len(cardLinks) == 0 {
			description = "Some card images could be found, nor could we find a MarvelCDB link."
			for _, c := range cardsWithErrors {
				for _, face := range c.Faces {
					link := fmt.Sprintf("%s, %s - %s", face.Name, c.Packs[0].SKU, c.Packs[0].Name)
					cardLinks = append(cardLinks, link)
				}
			}
		}
		fieldValue := strings.Join(cardLinks, "\r\n")

		// We need a field with all the card links
		field := &discordgo.MessageEmbedField{
			Name:   "Card Links:",
			Value:  fieldValue,
			Inline: false,
		}
		fields = append(fields, field)
	}

	// TODO - Make this a function as well
	// Return the image to Discord
	if horizontalFileName != "" {
		ms := &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Color:       0x78141b,
				Description: description,
				Fields:      fields,
				Image: &discordgo.MessageEmbedImage{
					URL: "attachment://" + horizontalFileName,
				},
			},
			Files: []*discordgo.File{
				{
					Name:        horizontalFileName,
					Reader:      horizontalFile,
					ContentType: "image/jpeg",
				},
			},
		}

		_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
		if err != nil {
			logError.Errorf("error sending message: %v", err)
		}
	}

	if verticalFileName != "" {
		ms := &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Color:       0x78141b,
				Description: description,
				Fields:      fields,
				Image: &discordgo.MessageEmbedImage{
					URL: "attachment://" + verticalFileName,
				},
			},
			Files: []*discordgo.File{
				{
					Name:        verticalFileName,
					Reader:      verticalFile,
					ContentType: "image/jpeg",
				},
			},
		}

		_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
		if err != nil {
			logError.Errorf("error sending message: %v", err)
		}
	}

	// Delete the new image file so we don't (eventually) run out of disk space
	if horizontalFileName != "" {
		err := os.Remove("images/" + horizontalFileName)
		if err != nil {
			logError.Errorf("error deleting file: %v", err)
		}
	}
	if verticalFileName != "" {
		err := os.Remove("images/" + verticalFileName)
		if err != nil {
			logError.Errorf("error deleting file: %v", err)
		}
	}
}

// sendRulesMessages will send an embedded Rule to the channel for each object in the rules slice
func sendRulesMessages(srv *Server, s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, rules []*rule.Rule) {
	// Configure logger for failed Discord message sends
	logError := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
		"user":   u.Username,
		"level":  "error",
	})

	for _, r := range rules {
		// Create our embed fields
		fields := []*discordgo.MessageEmbedField{}
		// If rules exceed 2048 characters, they won't fit in the Description field
		// So we'll parse those into extra fields
		if len(r.Text2) > 0 {
			rules_cont := &discordgo.MessageEmbedField{
				Name:  "Rules (continued)",
				Value: r.Text2,
			}
			fields = append(fields, rules_cont)
		}
		// If there is a "See also" section, append that as well
		if len(r.Related) > 0 {
			related_field := &discordgo.MessageEmbedField{
				Name:  "See also",
				Value: strings.Join(r.Related, "\n"),
			}
			fields = append(fields, related_field)
		}

		ms := &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Title:       r.Name,
				Description: r.Text, // Description allows 2048 characters
				Fields:      fields, // Embed fields allow 1024 characters
			},
		}
		_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
		if err != nil {
			logError.Errorf("error sending message: %v", err)
		}
	}
	return
}

// findCards is a function that takes a filter and a query string and returns the closest matching Cards.
func findCards(filter string, query string, cards []*card.Card) (matches []*card.Card) {
	switch filter {
	// Pack has different logic than the other filters, since we are not looking at the Type field
	case "pack":
		lowercaseMatches := []*card.Card{}
		containsMatches := []*card.Card{}
		levenshteinMatches := []*card.Card{}
		for _, c := range cards {
			for _, pack := range c.Packs {
				packName := strings.ToLower(pack.Name)
				// Pack name is an exact match
				// e.g., "captain america" == "captain america"
				if query == packName {
					lowercaseMatches = append(lowercaseMatches, c)
					break
				}
				// Set name is a contains match
				if strings.Contains(packName, query) {
					containsMatches = append(containsMatches, c)
					break
				}
			}
		}
		// Prefer exact card name matching
		if len(lowercaseMatches) > 0 {
			matches = append(matches, lowercaseMatches...)
			return
		}
		// Next, prefer "contains" card name matching
		if len(containsMatches) > 0 {
			matches = append(matches, containsMatches...)
			return
		}
		// If the other algorithms haven't matched, we'll use Levenshtein distance
		// This is in a separate for loop since we have to create a new object for each card
		levenshteinMatches = append(levenshteinMatches, findLevenshteinCards(filter, query, cards)...)
		matches = append(matches, levenshteinMatches...)
		return matches
	// Set has different logic than the other filters, since we are not looking at the Type field
	case "set":
		lowercaseMatches := []*card.Card{}
		containsMatches := []*card.Card{}
		levenshteinMatches := []*card.Card{}
		for _, c := range cards {
			for _, set := range c.Sets {
				setName := strings.ToLower(set.Name)
				// Set name is an exact match
				// e.g., "expert" == "expert"
				if query == setName {
					lowercaseMatches = append(lowercaseMatches, c)
					break
				}
				// Set name is a contains match
				if strings.Contains(setName, query) {
					containsMatches = append(containsMatches, c)
					break
				}
			}
		}
		// Prefer exact card name matching
		if len(lowercaseMatches) > 0 {
			matches = append(matches, lowercaseMatches...)
			return
		}
		// Next, prefer "contains" card name matching
		if len(containsMatches) > 0 {
			matches = append(matches, containsMatches...)
			return
		}
		// If the other algorithms haven't matched, we'll use Levenshtein distance
		// This is in a separate for loop since we have to create a new object for each card
		levenshteinMatches = append(levenshteinMatches, findLevenshteinCards(filter, query, cards)...)
		matches = append(matches, levenshteinMatches...)
		return matches
	// These filters compare against face.Type
	case "attachment", "ally", "alter-ego", "hero", "minion", "upgrade", "obligation", "support", "villain":
		lowercaseMatches := []*card.Card{}
		containsMatches := []*card.Card{}
		levenshteinMatches := []*card.Card{}
		for _, c := range cards {
			for _, cardName := range c.Names {
				// Lowercased name is an exact match
				// e.g., "peter parker" == "peter parker"
				cardName = strings.ToLower(cardName)
				if query == cardName {
					// Now we need to compare against the filter
					for _, cardFace := range c.Faces {
						if filter == strings.ToLower(cardFace.Type) {
							lowercaseMatches = append(lowercaseMatches, c)
							// Fix for Ant-Man and Wasp with "hero" filter
							// TODO - does adding a break here negatively impact any other cards?
							break
						}
					}
				}
				// Name contains the query string
				// eg, "hawkeye's quiver" == "quiver"
				if strings.Contains(cardName, query) {
					// Now we need to compare against the filter
					for _, cardFace := range c.Faces {
						if filter == strings.ToLower(cardFace.Type) {
							containsMatches = append(containsMatches, c)
						}
					}
				}
			}
		}
		// Prefer exact card name matching
		if len(lowercaseMatches) > 0 {
			matches = append(matches, lowercaseMatches...)
			return
		}
		// Next, prefer "contains" card name matching
		if len(containsMatches) > 0 {
			matches = append(matches, containsMatches...)
			return
		}
		// If the other algorithms haven't matched, we'll use Levenshtein distance
		// This is in a separate for loop since we have to create a new object for each card
		levenshteinMatches = append(levenshteinMatches, findLevenshteinCards(filter, query, cards)...)
		matches = append(matches, levenshteinMatches...)
		return matches
		// No filter, or the filter was not recognized
	default:
		lowercaseMatches := []*card.Card{}
		containsMatches := []*card.Card{}
		levenshteinMatches := []*card.Card{}
		for _, c := range cards {
			for _, cardName := range c.Names {
				// Lowercased name is an exact match
				// e.g., "peter parker" == "peter parker"
				cardName = strings.ToLower(cardName)
				if query == cardName {
					lowercaseMatches = append(lowercaseMatches, c)
				}
				// Name contains the query string
				// eg, "hawkeye's quiver" == "quiver"
				if strings.Contains(cardName, query) {
					containsMatches = append(containsMatches, c)
				}
			}
		}
		// Prefer exact card name matching
		if len(lowercaseMatches) > 0 {
			matches = append(matches, lowercaseMatches...)
			return
		}
		// Next, prefer "contains" card name matching
		if len(containsMatches) > 0 {
			matches = append(matches, containsMatches...)
			return
		}
		// If the other algorithms haven't matched, we'll use Levenshtein distance
		// This is in a separate for loop since we have to create a new object for each card
		levenshteinMatches = append(levenshteinMatches, findLevenshteinCards(filter, query, cards)...)
		matches = append(matches, levenshteinMatches...)
		return matches
	}
	return matches
}

// findLevenshteinCards is a function that returns the closest matching Cards by Levenshtein distance.
func findLevenshteinCards(filter string, query string, cards []*card.Card) (bestMatches []*card.Card) {
	// New type to hold the card, distance, and ratio
	type LevenshteinCard struct {
		card     *card.Card
		distance int
		ratio    float64
	}

	// Iterate through each card and calculate the Levenshtein distance and ratio
	options := levenshtein.DefaultOptions
	matches := []*LevenshteinCard{}
	for _, c := range cards {
		// Pack match logic
		if filter == "pack" {
			for _, pack := range c.Packs {
				packName := strings.ToLower(pack.Name)
				distance := levenshtein.DistanceForStrings([]rune(query), []rune(packName), options)
				ratio := levenshtein.RatioForStrings([]rune(query), []rune(packName), options)
				// 70% match is our cutoff point, we'll add these to a slice
				if ratio > 0.70 {
					matches = append(matches, &LevenshteinCard{c, distance, ratio})
				}
			}
			continue
		}
		// Set match logic
		if filter == "set" {
			for _, set := range c.Sets {
				setName := strings.ToLower(set.Name)
				distance := levenshtein.DistanceForStrings([]rune(query), []rune(setName), options)
				ratio := levenshtein.RatioForStrings([]rune(query), []rune(setName), options)
				// 70% match is our cutoff point, we'll add these to a slice
				if ratio > 0.70 {
					matches = append(matches, &LevenshteinCard{c, distance, ratio})
				}
			}
			continue
		}
		// Card name match logic
		for _, cardName := range c.Names {
			cardName = strings.ToLower(cardName)
			distance := levenshtein.DistanceForStrings([]rune(query), []rune(cardName), options)
			ratio := levenshtein.RatioForStrings([]rune(query), []rune(cardName), options)

			// 70% match is our cutoff point, we'll add these to a slice
			if ratio > 0.70 {
				// Compare against the filter - currently this only works for Face.Type
				// It is likely we'll need a different function for Sets, etc
				if filter != "" {
					for _, cardFace := range c.Faces {
						if filter == strings.ToLower(cardFace.Type) {
							matches = append(matches, &LevenshteinCard{c, distance, ratio})
						}
					}
				} else {
					matches = append(matches, &LevenshteinCard{c, distance, ratio})
				}
			}
		}
	}

	// We'll find the highest ratio match
	max := 0.0
	for _, c := range matches {
		if c.ratio > max {
			max = c.ratio
		}
	}
	// Return all cards with that ratio
	for _, c := range matches {
		if c.ratio == max {
			bestMatches = append(bestMatches, c.card)
		}
	}
	return
}

// findRule is a function that takes a query string and returns the closest matching Rule.
// TODO - Should rules be a map instead of a slice? Unlike cards, which can share a
// TODO - name without being unique (see: Wakanda Forever!), a rule name is useful as a key.
func findRule(query string, rules []*rule.Rule) *rule.Rule {
	for _, r := range rules {
		if query == strings.ToLower(r.Name) {
			return r
		}
	}
	// TODO - implement Levenshtein or similar fuzzy matching or search algorithm before returning a failure
	return nil
}

// trimCommand takes a bot command and removes the control brackets
// e.g., [[Lockjaw]] becomes Lockjaw
func trimCommand(cmd string) string {
	cmd = strings.TrimLeft(cmd, "[[")
	cmd = strings.TrimRight(cmd, "]]")
	return cmd
}
