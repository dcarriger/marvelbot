package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"marvelbot/card"
	"marvelbot/manualcards"
	"marvelbot/rule"
	"marvelbot/server"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Variables used for command line parameters
var (
	Token string
	Cards []*card.Card
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create a new Server.
	srv := server.NewServer(Token)

	// Get all cards from MarvelCDB.com
	err := srv.GetCards()
	if err != nil {
		srv.Logger.Fatalf("error fetching cards: %w", err)
	}
	// Append the status cards to the slice
	srv.Cards = append(srv.Cards, card.StatusCards...)
	// TODO - this should probably happen elsewhere
	// Override broken MarvelCDB images for some cards
	for _, card := range srv.Cards {
		// If a card isn't double-sided
		if strings.HasPrefix(card.Code, "16") && card.DoubleSided == false {
			card.ImageSrc = fmt.Sprintf("images/%s.png", card.Code)
		}
		// If a card is double-sided
		if strings.HasPrefix(card.Code, "16") && card.DoubleSided == true {
			card.ImageSrc = fmt.Sprintf("images/%sa.png", card.Code)
			card.BackImageSrc = fmt.Sprintf("images/%sb.png", card.Code)
		}
	}
	// Append any manual cards to the slice
	srv.Cards = append(srv.Cards, manualcards.ManualCards...)

	// Read all of the rule files
	files, err := ioutil.ReadDir("rules/")
	if err != nil {
		srv.Logger.Fatalf("error fetching rules: %w", err)
	}
	for _, f := range files {
		r := &rule.Rule{}
		yamlFile, err := ioutil.ReadFile(fmt.Sprintf("rules/%s", f.Name()))
		if err != nil {
			srv.Logger.Errorf("error reading rules file %v: %w", f.Name(), err)
			continue
		}
		err = yaml.Unmarshal(yamlFile, r)
		if err != nil {
			srv.Logger.Errorf("error unmarshaling yaml file %v: %w", f.Name(), err)
			continue
		}
		srv.Rules = append(srv.Rules, r)
	}

	// Register the MessageCreate func as a callback for MessageCreate events.
	srv.Session.AddHandler(srv.MessageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = srv.Session.Open()
	if err != nil {
		srv.Logger.Fatal("error opening Discord websocket: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	srv.Logger.Info("Bot started.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	srv.Session.Close()
}
