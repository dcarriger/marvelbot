package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
	"marvelbot/pkg/card"
	"marvelbot/pkg/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// Register the MessageCreate func as a callback for MessageCreate events.
	srv.Session.AddHandler(srv.MessageCreate)

	// Add handlers for all of our slash commands
	srv.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := srv.Handlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = srv.Session.Open()
	if err != nil {
		srv.Logger.Fatal("error opening Discord websocket: ", err)
		return
	}

	/*
		// Remove all global commands we accidentally registered
		cmd, err := srv.Session.ApplicationCommands(srv.Session.State.User.ID, "")
		fmt.Println(len(cmd))
		if err != nil {
			srv.Logger.Fatal("Unable to fetch global commands")
		}
		for _, v := range cmd {
			err := srv.Session.ApplicationCommandDelete(srv.Session.State.User.ID, "", v.ID)
			if err != nil {
				srv.Logger.Fatal("Unable to delete command")
			}
		}
		fmt.Println("all global commands removed!")

		// Remove all Guild commands we accidentally registered
		guildCmd, err := srv.Session.ApplicationCommands(srv.Session.State.User.ID, "671913936576053289")
		fmt.Println(len(guildCmd))
		if err != nil {
			srv.Logger.Fatal("Unable to fetch Guild commands")
		}
		for _, v := range guildCmd {
			fmt.Println(v.ID, v.Description, v.Name)
			err := srv.Session.ApplicationCommandDelete(srv.Session.State.User.ID, "671913936576053289", v.ID)
			if err != nil {
				srv.Logger.Fatal("Unable to delete command")
			}
		}
		fmt.Println("all Guild commands removed!")
	*/

	// Register commands to the Guild (test guild, currently)
	// Dev - 671913936576053289
	// MC - 607399156780105741
	for _, v := range srv.Commands {
		_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, "671913936576053289", v)
		if err != nil {
			srv.Logger.Fatal(fmt.Sprintf("Cannot create '%v' command: %v", v.Name, err))
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	srv.Logger.Info("Bot started.")
	time.Sleep(2 * time.Second)
	for _, guild := range srv.Session.State.Guilds {
		srv.Logger.Info(fmt.Sprintf("Joined: %s - %s - %s", guild.ID, guild.Name, guild.Description))
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	srv.Session.Close()
}
