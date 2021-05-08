package main

import (
	"flag"
	"fmt"
	"marvelbot/server"
	"os"
	"os/signal"
	"syscall"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create a new Server.
	srv := server.NewServer(Token)

	// Download all card images
	/*
		for _, c := range srv.Cards {
			err := c.DownloadImages()
			if err != nil {
				srv.Logger.Error(err)
			}
		}
	*/

	/*
		// Get MarvelCDB cards
		const baseURL = "https://marvelcdb.com/api/public"
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/cards/?_format=json&encounter=1", baseURL), nil)
		if err != nil {
			fmt.Println("failed to build request:", err)
		}

		client := &http.Client{
			Timeout: time.Second * 30,
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("request failed:", err)
		}
		// TODO - handle error closing the response?
		defer resp.Body.Close()

		// Create a new slice to hold the unmarshaled response
		mcdbCards := []*card.MarvelCDBCard{}

		// Unmarshal the JSON response into individual cards
		if err := json.NewDecoder(resp.Body).Decode(&mcdbCards); err != nil {
			fmt.Println("error unmarshaling:", err)
		}

		// Convert each card
		convertedCards := []*card.Card{}
		for _, mcdbCard := range mcdbCards {
			converted := mcdbCard.Convert()
			convertedCards = append(convertedCards, converted)
		}

		// Write the converted YAML to a file somewhere
		yamlText, err := yaml.Marshal(&convertedCards)
		f, err := os.Create("data/converted.yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		_, err = f.Write(yamlText)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/

	/*
		// Convert manual cards over to our YAML format
		var manualCards []*card.Card
		for _, c := range manualcards.ManualCards {
			convertedCard := c.Convert()
			manualCards = append(manualCards, convertedCard)
		}

		// Write the converted YAML to a file somewhere
		yamlText, err := yaml.Marshal(&manualCards)
		f, err := os.Create("data/manual.yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		_, err = f.Write(yamlText)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/

	// Register the MessageCreate func as a callback for MessageCreate events.
	srv.Session.AddHandler(srv.MessageCreate)

	// Open a websocket connection to Discord and begin listening.
	err := srv.Session.Open()
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
