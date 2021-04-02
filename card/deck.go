package card

// DeckList represents the JSON response we get from MarvelCDB. Cards are represented as a map of card IDs and their
// quantities.
type DeckList struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	DateCreated string         `json:"date_creation"`
	DateUpdated string         `json:"date_update"`
	Description string         `json:"description_md"`
	UserID      int            `json:"user_id"`
	HeroCode    string         `json:"investigator_code"`
	HeroName    string         `json:"investigator_name"`
	Slots       map[string]int `json:"slots"`
	Version     string         `json:"version"`
	Meta        string         `json:"meta"`
}

// Deck represents a Marvel Champions deck, with the card IDs substituted for Card objects.
type Deck struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	DateCreated string        `json:"date_creation"`
	DateUpdated string        `json:"date_update"`
	Description string        `json:"description_md"`
	UserID      int           `json:"user_id"`
	HeroCode    string        `json:"hero_code"`
	HeroName    string        `json:"hero_name"`
	Cards       map[*Card]int `json:"cards"`
	Version     string        `json:"version"`
	Meta        string        `json:"meta"`
}

type Meta struct {
	Aspect string `json:"aspect"`
}

// NewDeck builds a Deck from a MarvelCDB DeckList.
func NewDeck(deckList *DeckList, cardList []*Card) *Deck {
	deck := &Deck{
		ID:          deckList.ID,
		Name:        deckList.Name,
		DateCreated: deckList.DateCreated,
		DateUpdated: deckList.DateUpdated,
		Description: deckList.Description,
		UserID:      deckList.UserID,
		HeroCode:    deckList.HeroCode,
		HeroName:    deckList.HeroName,
		Cards:       parseCards(deckList.Slots, cardList),
		Version:     deckList.Version,
		Meta:        deckList.Meta,
	}
	return deck
}

func parseCards(cardMap map[string]int, cardList []*Card) map[*Card]int {
	// Range over the full list of cards and check for intersections with the list
	cards := make(map[*Card]int)
	for _, card := range cardList {
		if quantity, ok := cardMap[card.Code]; ok {
			cards[card] = quantity
		}
	}
	return cards
}
