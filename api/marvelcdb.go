package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"marvelbot/card"
	"net/http"
)

// GetDeckList fetches the deck ID from the MarvelCDB API and parses it into a DeckList.
func GetDeckList(deckID string) (deckList *card.DeckList, err error) {
	resp, err := http.Get("https://marvelcdb.com/api/public/decklist/" + deckID + ".json")
	if err != nil {
		return nil, fmt.Errorf("Error fetching deck list %s from MarvelCDB: %w", deckID, err)
	}
	defer resp.Body.Close()

	// If we got an empty response, then the deck doesn't exist
	if resp.ContentLength == 0 {
		return nil, fmt.Errorf("No deck found with id %s", deckID)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body from MarvelCDB for deck id %s: %w", deckID, err)
	}

	// Unmarshal the response body into a DeckList
	err = json.Unmarshal(body, &deckList)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON response from MarvelCDB: %w", err)
	}

	return deckList, nil
}
