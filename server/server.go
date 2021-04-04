package server

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"marvelbot/card"
	"marvelbot/rule"
	"net/http"
	"os"
	"time"
)

// The base URL of the MarvelCDB API.
const baseURL = "https://marvelcdb.com/api/public"

// Server is used to handle dependency injection into our bot.
type Server struct {
	Session *discordgo.Session
	Client  *http.Client
	Cards   []*card.Card
	Rules   []*rule.Rule
	Logger  *logrus.Logger
}

// NewServer creates a new Server with all our expected dependencies - cards, logging, and a working Discord session.
func NewServer(token string) (s *Server) {
	// Create a new Logger
	log := logrus.New()

	// Open our log file
	file, err := os.OpenFile("./bot.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Configure our output and default log level
	log.SetOutput(file)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session: ", err)
		return
	}

	// Create an HTTP client for use with external HTTP calls
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// This is initially empty - it's up to the caller to populate it.
	cards := []*card.Card{}

	// Build and return our server
	s = &Server{
		Session: dg,
		Client:  client,
		Cards:   cards,
		Logger:  log,
	}
	return s
}

// GetCards fetches all cards from MarvelCDB and updates the Server object.
func (s *Server) GetCards() (err error) {
	// Build our GET request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/cards/?_format=json&encounter=1", baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Make our GET request
	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	// TODO - handle error closing the response?
	defer resp.Body.Close()

	// Create a new slice to hold the unmarshaled response
	cards := []*card.Card{}

	// Unmarshal the JSON response into individual cards
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return fmt.Errorf("error unmarshaling: %w", err)
	}

	// Update the server struct with the response and return
	s.Cards = cards
	return nil
}
