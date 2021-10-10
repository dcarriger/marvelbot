package server

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"marvelbot/pkg/card"
	"marvelbot/pkg/rule"
	"net/http"
	"os"
	"time"
)

// Server is used to handle dependency injection into our bot.
type Server struct {
	Session  *discordgo.Session
	Client   *http.Client
	Commands []*discordgo.ApplicationCommand
	Handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Cards    []*card.Card
	Homebrew []*card.Card
	Rules    []*rule.Rule
	Logger   *logrus.Logger
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

	// Read data in from our folder containing card YAML data
	cards, err := ReadCards("data/cards")
	if err != nil {
		log.Fatal("error reading card data: ", err)
	}

	// Read data in from our folder containing homebrew card YAML data
	homebrew, err := ReadCards("data/homebrew")
	if err != nil {
		log.Fatal("error reading homebrew card data: ", err)
	}

	// Read data in from our folder containing rule YAML data
	rules, err := ReadRules("data/rules")
	if err != nil {
		log.Fatal("error reading rules data: ", err)
	}

	// Build and return our server
	s = &Server{
		Session:  dg,
		Client:   client,
		Commands: commands,
		Cards:    cards,
		Homebrew: homebrew,
		Rules:    rules,
		Logger:   log,
	}

	// Append our handlers (which need access to the Cards object inside the Server)
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"card":    s.CardHandler,
		"mission": s.MissionHandler,
	}
	s.Handlers = handlers

	return s
}
