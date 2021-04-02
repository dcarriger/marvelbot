package server

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	gim "github.com/ozankasikci/go-image-merge"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	_ "image/jpeg"
	"image/png"
	"marvelbot/api"
	"marvelbot/card"
	"marvelbot/rule"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// These constants map to the color codes used by different Aspects.
const (
	Aggression = 0x78141b
	Basic      = 0x8c8c8c
	Justice    = 0xa09320
	Leadership = 0x3ea0b2
	Protection = 0x59aa36
)

// This function will be called every time a new message is created on any channel that the authenticated bot has
// access to (due to the DiscordGo AddHandler).
func (srv *Server) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	// Help utilizing the Discord bot
	if m.Content == "!help" {
		srv.HandleHelp(s, m)
	}

	// Get decklist for a given deck
	deckRegexp := regexp.MustCompile(`^!deck [0-9]+$`)
	if deckRegexp.MatchString(m.Content) {
		srv.HandleDeckList(s, m)
	}

	// Look for occurrences of '[[card]]'
	cardRegexp := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	// We found a match
	if cardRegexp.MatchString(m.Content) {
		srv.HandleCards(s, m)
	}
}

// TODO - some cards have a back_img_src, need to support that as well both in the Cards struct and when creating card images

// HandleHelp is a function that returns supported bot commands.
func (srv *Server) HandleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Configure a logger for failed Discord message sends
	logFailure := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
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

	fetchDeck := &discordgo.MessageEmbedField{
		Name: "!deck <deck id>",
		Value: "Fetches the deck list for <deck id> from MarvelCDB.com and outputs it for easy reference. Decks " +
			"must be published on MarvelCDB for the bot to access them. If a deck is unpublished, it will not " +
			"be visible to the API and thus the bot will not find it.",
		Inline: false,
	}

	displayHelp := &discordgo.MessageEmbedField{
		Name:   "!help",
		Value:  "Displays this help message.",
		Inline: false,
	}

	fields = append(fields, fetchCard, fetchDeck, displayHelp)

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
		logFailure.Errorf("error sending message: %w", err)
	}
	return
}

// HandleCards is a function that looks for cards in the format of "[[card name]]", parses the results, and returns
// them to the channel.
func (srv *Server) HandleCards(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Configure logger for failed Discord message sends
	logFailure := srv.Logger.WithFields(log.Fields{
		"server": m.GuildID,
	})

	cardRegexp := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	// Let's find out how many matches we have and iterate through them
	results := cardRegexp.FindAllString(m.Content, -1)

	// Pass all cards to the findMatches endpoint
	matchingCards := []*card.Card{}
	rules := []*rule.Rule{}
	for _, v := range results {
		// Initialize default values
		name := v
		filter := ""
		// Check if there is a filter -
		if strings.Contains(v, ":") {
			splitStr := strings.Split(v, ":")
			filter = normalize(splitStr[0])
			name = normalize(splitStr[1])
		}
		// TODO - I wrote this handler before I was supporting rules
		// It may be better to write a different handler that checks
		// whether we're looking for cards or rules, then passes the
		// value to some other function to handle it.
		if filter == "rule" {
			found := findRule(name, srv.Rules)
			if found != nil {
				rules = append(rules, found)
			}
		}
		found := findMatches(name, filter, srv.Cards, true)
		matchingCards = append(matchingCards, found...)
	}

	// We found rules - let's return the rules and exit
	// TODO - Currently, returning a mix of rules and cards is unsupported
	// Maybe figure out how to handle that case?
	if len(rules) > 0 {
		for _, r := range rules {
			// Create our embed fields
			fields := []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Rule Text",
					Value:  r.Text,
					Inline: true,
				},
			}
			ms := &discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Title: r.Name,
					Fields: fields,
				},
			}
			_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
			if err != nil {
				logFailure.Errorf("error sending message: %v", err)
			}
			return
		}
		msg := fmt.Sprintf("Found a matching rule: %s", m.Content)
		_, err := s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			logFailure.Errorf("error sending message: %w", err)
		}
		return
	}

	// No results were found
	if len(matchingCards) == 0 {
		/*
			// TODO - doesn't work great with multiple non-matching cards
			msg := fmt.Sprintf("No matching cards found: %s", strings.TrimFunc(m.Content, func(r rune) bool {
				if r == '[' || r == ']' {
					return true
				}
				return false
			}))
		*/
		msg := fmt.Sprintf("No matching cards found: %s", m.Content)
		_, err := s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			logFailure.Errorf("error sending message: %w", err)
		}
		return
	}

	// Build our Grid to return images
	// https://github.com/ozankasikci/go-image-merge#examples
	grids := []*gim.Grid{}

	// Find out if we need to rotate schemes or not
	var rotateSchemes bool
	for _, c := range matchingCards {
		if scheme := c.IsScheme(); scheme == false {
			rotateSchemes = true
		}
	}

	// If we have any cards missing images, we'll use this to return a link to the card
	cardsWithErrors := []*card.Card{}

	// Loop through all of our Cards and add matching images to the Grid
	for _, c := range matchingCards {
		// Download the card images if we don't already have them
		err := c.DownloadImages()
		if err != nil {
			logFailure.Errorf("failed to download images: %w", err)
			cardsWithErrors = append(cardsWithErrors, c)
			continue
		}

		// Get all image paths
		var imagePaths []string
		images := c.GetImages()
		scheme := c.IsScheme()
		for _, imageURL := range images {
			var path string
			if rotateSchemes == true && scheme == true {
				path = card.GetImagePath(imageURL, true)
			} else {
				path = card.GetImagePath(imageURL, false)
			}
			imagePaths = append(imagePaths, path)
		}

		// Add everything to the Grid now
		for _, value := range imagePaths {
			grid := &gim.Grid{
				ImageFilePath: value,
			}
			grids = append(grids, grid)
		}
	}

	// Let's make sure our Grid isn't empty - if it is, we need to send the card links to the channel
	if len(grids) == 0 {
		// Get all the card links
		var cardLinks []string

		// Build the return string
		for _, c := range matchingCards {
			link := fmt.Sprintf("[%s](%s)", c.Name, c.URL)
			cardLinks = append(cardLinks, link)
		}
		fieldValue := strings.Join(cardLinks, "\r\n")

		// We need a field with all the card links
		fields := []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Card Links:",
				Value:  fieldValue,
				Inline: false,
			},
		}

		// Build the return message
		ms := &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Author:      &discordgo.MessageEmbedAuthor{},
				Color:       Basic,
				Description: "No card images were found.",
				Fields:      fields,
				Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			},
		}

		// Send the message to the channel
		_, err := s.ChannelMessageSendComplex(m.ChannelID, ms)
		if err != nil {
			logFailure.Errorf("error sending message: %w", err)
		}
		return
	}

	// Return the Discord message
	// Max 3 cards per row
	var horizontalCount, verticalCount int
	if len(grids) > 3 {
		horizontalCount = 3
		verticalCount = int(math.Ceil(float64(len(grids)) / 3.0))
	} else {
		horizontalCount = len(grids)
		verticalCount = 1
	}

	// Merge the images into a grid
	rgba, err := gim.New(grids, horizontalCount, verticalCount).Merge()
	if err != nil {
		logFailure.Errorf("error merging images: %w", err)
	}

	// Create a guid to use for our new image
	guid := ksuid.New()
	fileName := guid.String() + ".png"

	// save the output to jpg or png
	file, err := os.Create("images/" + fileName)
	if err != nil {
		logFailure.Errorf("error saving file: %w", err)
	}
	err = png.Encode(file, rgba)
	if err != nil {
		logFailure.Errorf("error encoding PNG: %w", err)
	}
	defer file.Close()

	// Rewind to beginning of file after writing to it
	_, err = file.Seek(0, 0)
	if err != nil {
		msg := fmt.Sprintf("Error rewinding file: %v", err)
		_, err = s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			logFailure.Errorf("error sending message: %w", err)
		}
		return
	}

	// Did we have any errors? If so, we'll return them as fields
	var description string
	var fields []*discordgo.MessageEmbedField

	if len(cardsWithErrors) != 0 {
		description = "Some card images could not be returned - these are linked below."
		// Get all the card links
		var cardLinks []string

		// Build the return string
		for _, c := range cardsWithErrors {
			link := fmt.Sprintf("[%s](%s)", c.Name, c.URL)
			cardLinks = append(cardLinks, link)
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

	// Return the new image to Discord
	ms := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:       0x78141b,
			Description: description,
			Fields:      fields,
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://" + fileName,
			},
		},
		Files: []*discordgo.File{
			{
				Name:        fileName,
				Reader:      file,
				ContentType: "image/jpeg",
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, ms)
	if err != nil {
		logFailure.Errorf("error sending message: %w", err)
	}

	// Delete the new image file so we don't (eventually) run out of disk space
	err = os.Remove("images/" + fileName)
	if err != nil {
		logFailure.Errorf("error deleting file: %w", err)
	}
}

func (srv *Server) HandleDeckList(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get the deck number to fetch from the Discord message
	// NOTE - this should always work otherwise the regex wouldn't have matched
	deckID := strings.Split(m.Content, " ")[1]

	// Query MarvelCDB
	deckList, err := api.GetDeckList(deckID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprint(err))
		return
	}

	// Parse the cards into a new Deck
	deck := card.NewDeck(deckList, srv.Cards)

	// Parse the Aspect from the deck
	aspect := &card.Meta{}
	err = json.Unmarshal([]byte(deck.Meta), aspect)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error unmarshaling Aspect: %v", err))
		return
	}

	// Aspect-based color picker
	var color int
	switch aspect.Aspect {
	case "aggression":
		color = Aggression
	case "justice":
		color = Justice
	case "leadership":
		color = Leadership
	case "protection":
		color = Protection
	}

	// Create our embed fields
	// We'll do one each for Hero, Aspect, and Basic cards
	fields := []*discordgo.MessageEmbedField{}

	// We want the total card count, along with many other things - we'll range over the deck list once and pull out
	// everything we need
	count := 0
	energyResources := 0
	physicalResources := 0
	mentalResources := 0
	wildResources := 0
	heroCards := card.Cards{}
	aspectCards := card.Cards{}
	basicCards := card.Cards{}
	for item, quantity := range deck.Cards {
		// Increment card count
		count = count + quantity

		// Do all of our resource counts here
		energyResources = energyResources + (item.ResourceEnergy * quantity)
		mentalResources = mentalResources + (item.ResourceMental * quantity)
		physicalResources = physicalResources + (item.ResourcePhysical * quantity)
		wildResources = wildResources + (item.ResourceWild * quantity)

		// Add our card to the appropriate field string
		switch item.FactionCode {
		case "basic":
			basicCards = append(basicCards, item)
		case "hero":
			heroCards = append(heroCards, item)
		default:
			aspectCards = append(aspectCards, item)
		}
	}

	// Sort our card slices
	heroCards = heroCards.SortSlice()
	basicCards = basicCards.SortSlice()
	aspectCards = aspectCards.SortSlice()

	// We need string slices for our fields so we'll build those
	var heroCardsStr []string
	var aspectCardsStr []string
	var basicCardsStr []string

	for _, v := range heroCards {
		heroCardsStr = append(heroCardsStr, fmt.Sprintf("%s %s x%d", v.CostIcon(), v.EmbedString(), deck.Cards[v]))
	}

	for _, v := range aspectCards {
		aspectCardsStr = append(aspectCardsStr, fmt.Sprintf("%s %s x%d", v.CostIcon(), v.EmbedString(), deck.Cards[v]))
	}

	for _, v := range basicCards {
		basicCardsStr = append(basicCardsStr, fmt.Sprintf("%s %s x%d", v.CostIcon(), v.EmbedString(), deck.Cards[v]))
	}

	// Build our fields to return
	sl := []string{"Hero", "Aspect", "Basic"}
	for _, v := range sl {
		var myStr []string
		switch v {
		case "Hero":
			myStr = heroCardsStr
		case "Aspect":
			myStr = aspectCardsStr
		case "Basic":
			myStr = basicCardsStr
		}
		field := &discordgo.MessageEmbedField{
			Name:   v,
			Value:  strings.Join(myStr, "\n"),
			Inline: true,
		}
		fields = append(fields, field)
	}

	description := fmt.Sprintf("%s ● %s ● %d cards\n%d <:energy:665659193755303956> ● %d <:mental:665659107914809371> ● %d <:physical:665659151241969752> ● %d <:wild:665659016806137886>",
		deck.HeroName, strings.Title(aspect.Aspect), count, energyResources, mentalResources, physicalResources, wildResources,
	)

	// Set deck URL
	deckURL := "https://marvelcdb.com/decklist/view/" + deckID

	// Set thumbnail image
	thumbnail := ""
	switch deck.HeroName {
	case "Black Panther":
		thumbnail = "blackpanther.png"
	case "Black Widow":
		thumbnail = "blackwidow.png"
	case "Captain America":
		thumbnail = "captainamerica.png"
	case "Captain Marvel":
		thumbnail = "captainmarvel.png"
	case "Dr. Strange":
		thumbnail = "drstrange.png"
	case "Hulk":
		thumbnail = "hulk.png"
	case "Iron Man":
		thumbnail = "ironman.png"
	case "Ms. Marvel":
		thumbnail = "msmarvel.png"
	case "She-Hulk":
		thumbnail = "shehulk.png"
	case "Spider-Man":
		thumbnail = "spiderman.png"
	case "Thor":
		thumbnail = "thor.png"
		//default:
		//	thumbnailURL = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcS5q3Bb6Esp6Mq_CZe5nb70Gmq2r_FNdGCV9qbv_9umNOwQp9Ly&s"
	}

	// Open our attachment thumbnail
	fileName := "images/" + thumbnail
	f, err := os.Open(fileName)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprint(err))
		return
	}
	defer f.Close()

	// TODO - value limit cannot be above 1024 for fields
	ms := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			URL:         deckURL,
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       color,
			Description: description,
			Fields:      fields,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "attachment://" + thumbnail,
			},
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     deckList.Name,
		},
		Files: []*discordgo.File{
			{
				Name:        thumbnail,
				Reader:      f,
				ContentType: "image/jpeg",
			},
		},
	}
	_, err = s.ChannelMessageSendComplex(m.ChannelID, ms)
	return
}

func levenshteinCards(name string, cards []*card.Card) []*card.Card {
	// New type to hold the card, distance, and ratio
	type CardWithLev struct {
		card     *card.Card
		distance int
		ratio    float64
	}

	options := levenshtein.DefaultOptions
	// Iterate through each card and calculate the Levenshtein distance and ratio
	matches := []*CardWithLev{}
	for _, v := range cards {
		// Normalize the card names
		currentCard := v.Normalize()
		distance := levenshtein.DistanceForStrings([]rune(name), []rune(currentCard), options)
		ratio := levenshtein.RatioForStrings([]rune(name), []rune(currentCard), options)

		// 70% match is our cutoff point, we'll add these to a slice
		if ratio > 0.70 {
			matches = append(matches, &CardWithLev{v, distance, ratio})
		}
	}
	// Let's get the highest ratio match
	max := 0.0
	for _, v := range matches {
		if v.ratio > max {
			max = v.ratio
		}
	}

	// Return all cards with that ratio
	levCards := []*card.Card{}
	for _, v := range matches {
		if v.ratio == max {
			levCards = append(levCards, v.card)
		}
	}
	return levCards
}

// normalize is a function that takes a string and returns a normalized representation for card name matching.
func normalize(s string) string {
	s = strings.ToLower(s)
	// Remove punctuation so we don't have to deal with "Wakanda Forever!" or "Get Behind Me!" et al
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	s = re.ReplaceAllString(s, "")
	return s
}

// findRule is a function that takes a string and returns the closest matching Rule.
// TODO - Should rules be a map instead of a slice? Unlike cards, which can share a
// name without being unique (see: Wakanda Forever!), a rule name is useful as a key.
func findRule(name string, rules []*rule.Rule) *rule.Rule {
	name = normalize(name)
	for _, r := range rules {
		if name == normalize(r.Name) {
			return r
		}
	}
	return nil
}

// findMatches is a function that takes a string and optional filter and returns all matching Cards.
func findMatches(name string, filter string, Cards []*card.Card, fuzzyMatch bool) (matchingCards []*card.Card) {
	name = normalize(name)
	filter = normalize(filter)

	switch filter {
	// We are searching for all cards in a set, ie, "A Mess of Things", "Standard", "Expert"
	// TODO - refactor duplicated case statement logic into reusable function - DRY
	case "ally":
		matchCount := 0
		for _, c := range Cards {
			if c.TypeCode != "ally" {
				continue
			}
			currentCard := c.Name
			currentCard = normalize(currentCard)
			if name == currentCard {
				matchingCards = append(matchingCards, c)
				matchCount++
			}
		}
	case "hero":
		matchCount := 0
		for _, c := range Cards {
			if c.TypeCode != "hero" {
				continue
			}
			currentCard := c.Name
			currentCard = normalize(currentCard)
			if name == currentCard {
				matchingCards = append(matchingCards, c)
				matchCount++
			}
		}
	case "set":
		matchCount := 0
		for _, c := range Cards {
			currentCard := c.SetName
			currentCard = normalize(currentCard)
			if name == currentCard {
				matchingCards = append(matchingCards, c)
				matchCount++
			}
		}
	// If we haven't found anything, we can try again with fuzzy matching
	// TODO - make fuzzy matching work with set
	/*
		if matchCount == 0 && fuzzyMatch == true {
			matchingCards = levenshteinCards(name, Cards)
		}
	*/
	// No filter is applied
	case "":
		// Iterate over Cards to find matches
		matchCount := 0
		for _, c := range Cards {
			currentCard := c.Normalize()
			if name == currentCard {
				matchingCards = append(matchingCards, c)
				matchCount++
			}
		}
		// If we haven't found anything, we can try again with fuzzy matching
		if matchCount == 0 && fuzzyMatch == true {
			matchingCards = levenshteinCards(name, Cards)
		}
	// A filter is applied that we don't recognize
	default:
		// Iterate over Cards to find matches
		matchCount := 0
		for _, c := range Cards {
			currentCard := c.Normalize()
			if name == currentCard {
				matchingCards = append(matchingCards, c)
				matchCount++
			}
		}
		// If we haven't found anything, we can try again with fuzzy matching
		if matchCount == 0 && fuzzyMatch == true {
			matchingCards = levenshteinCards(name, Cards)
		}
	}
	return matchingCards
}

// imagePath takes a non-local image path (such as a MarvelCDB path) and converts it into our local path.
func imagePath(path string) string {
	// Empty URL
	if path == "" {
		return ""
	}

	// Split the path into parts
	imageSlice := strings.Split(path, "/")
	imageName := strings.ToLower(imageSlice[len(imageSlice)-1])

	// Remove the extension from the path
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName))

	// We want to save images as PNG so we'll define our image path here
	imagePath := "images/" + imageName + ".png"

	return imagePath
}
