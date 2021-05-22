package server

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"marvelbot/card"
	"marvelbot/rule"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The base URL of the MarvelCDB API.
const baseURL = "https://marvelcdb.com/api/public"

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
	cards, err := ReadCards("data/")
	if err != nil {
		log.Fatal("error reading card data: ", err)
	}

	// Read data in from our folder containing homebrew card YAML data
	homebrew, err := ReadCards("homebrew/")
	if err != nil {
		log.Fatal("error reading homebrew card data: ", err)
	}

	// Read data in from our folder containing rule YAML data
	rules, err := ReadRules("rules/")
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
		"card": s.CardHandler,
	}
	s.Handlers = handlers

	return s
}

// ReadRules parses rules in YAML format and returns them to the caller.
func ReadRules(path string) (rules []*rule.Rule, err error) {
	// Read the list of YAML files in the path
	// TODO - make this only list YAML files and not all files
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", path, err)
	}
	// Iterate over the files and unmarshal the underlying YAML data
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}
		unmarshaledRule := &rule.Rule{}
		yamlFile, err := ioutil.ReadFile(fmt.Sprintf("rules/%s", f.Name()))
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s: %w", f.Name(), err)
		}
		err = yaml.Unmarshal(yamlFile, &unmarshaledRule)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", f.Name(), err)
		}
		rules = append(rules, unmarshaledRule)
	}
	return
}

// ReadCards parses cards in YAML format and returns them to the caller.
func ReadCards(path string) (cards []*card.Card, err error) {
	// Read the list of YAML files in the path
	var files []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false && strings.HasSuffix(info.Name(), ".yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking %s: %w", path, err)
	}

	/*
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("unable to read %s: %w", path, err)
		}
	*/

	// Iterate over the files and unmarshal the underlying YAML data
	for _, f := range files {
		unmarshaledCards := []*card.Card{}
		yamlFile, err := ioutil.ReadFile(fmt.Sprintf("%s", f))
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s: %w", f, err)
		}
		err = yaml.Unmarshal(yamlFile, &unmarshaledCards)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", f, err)
		}
		cards = append(cards, unmarshaledCards...)
	}
	return
}
