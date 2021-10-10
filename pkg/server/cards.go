package server

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"marvelbot/pkg/card"
)

// ReadCards parses cards in YAML format and returns them to the caller.
func ReadCards(path string) (cards []*card.Card, err error) {
	files, err := walkConfigs(path, ".yaml")
	if err != nil {
		return nil, err
	}

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
