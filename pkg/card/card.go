package card

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const IMAGE_BASEDIR = "images"

// Card is a physical representation of a Marvel Champions card. Each instance of Card should represent the same
// physical copy of the card. For example, Peter Parker/Spider-Man is a single card. However, Rhino I, Rhino II, and
// Rhino III are three separate cards.
type Card struct {
	Names      []string                                      `json:"names" yaml:"names"`                     // Names in which the card is known by.
	Packs      []*Pack                                       `json:"packs,omitempty" yaml:"packs,omitempty"` // Packs in which the card has appeared.
	Sets       []*Set                                        `json:"sets,omitempty"  yaml:"sets,omitempty"`  // Sets in which the card is a member.
	Faces      []*Face                                       `json:"faces,omitempty" yaml:"faces,omitempty"` // Supports multiple sides of the card.
	*Deck      `json:"deck,omitempty" yaml:"deck,omitempty"` // Deck is a pointer to allow for things like status cards.
	Horizontal bool                                          `json:"horizontal,omitempty" yaml:"horizontal,omitempty"` // Whether the card is rotated horizontally.
}

// Deck represents which deck type the card belongs to.
type Deck int

const (
	Player = Deck(iota)
	Encounter
	Villain
	MainScheme
	Invocation
)

func (d Deck) String() string {
	name := []string{"Player", "Encounter", "Villain", "Main Scheme", "Invocation"}
	i := uint8(d)
	switch {
	case i <= uint8(Invocation):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

// Face represents a single card face. Most cards have only a single face. However, there are many cards that provide
// exceptions to this rule. Hero cards have an alter-ego and a hero side, and in the case of Ant-Man and Wasp, have
// multiple hero sides. Some villains, such as Wrecking Crew or Collector, have multiple sides as well. Some environment
// cards are also double-sided.
type Face struct {
	// The name of the card, e.g. Spider-Man
	Name string `json:"name" yaml:"name"`
	// The subtitle (for allies with subtitles), e.g., Miles Morales
	Subtitle *string `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	// The cost for a player to play the card
	Cost *int `json:"cost,omitempty" yaml:"cost,omitempty"`
	// Card type, such as Event, Upgrade, Ally, etc.
	Type string `json:"type" yaml:"type"`
	// Whether the card is unique
	Unique bool `json:"unique" yaml:"unique"`
	// Aspect is a slice to support future cards that may count as multiple Aspects
	Aspect []string `json:"aspect,omitempty" yaml:"aspect,omitempty"`
	// Basic REC value for Alter-Egos
	RecoverValue *int `json:"recover,omitempty" yaml:"recover,omitempty"`
	// Minion or Villain SCH value
	SchemeValue *int `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	// Any special text that triggers when scheming
	SchemeText *string `json:"scheme_text,omitempty" yaml:"scheme_text,omitempty"`
	// Basic THW value
	ThwartValue *int `json:"thwart,omitempty" yaml:"thwart,omitempty"`
	// Any special text that triggers when thwarting
	ThwartText *string `json:"thwart_text,omitempty" yaml:"thwart_text,omitempty"`
	// Consequential damage taken by an Ally when thwarting
	ThwartConsequential *int `json:"thwart_consequential,omitempty" yaml:"thwart_consequential,omitempty"`
	// Basic ATK value
	AttackValue *int `json:"attack,omitempty" yaml:"attack,omitempty"`
	// Any special text that triggers when attacking
	AttackText *string `json:"attack_text,omitempty" yaml:"attack_text,omitempty"`
	// Consequential damage taken by an Ally when attacking
	AttackConsequential *int `json:"attack_consequential,omitempty" yaml:"attack_consequential,omitempty"`
	// Basic DEF value
	DefenseValue *int `json:"defense,omitempty" yaml:"defense,omitempty"`
	// Any special text that triggers when defending
	DefenseText *string `json:"defense_text,omitempty" yaml:"defense_text,omitempty"`
	// The associated traits, e.g. S.H.I.E.L.D., Spy
	Traits []string `json:"traits,omitempty" yaml:"traits,omitempty"`
	// The associated keywords, e.g. Guard, Toughness
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	// The card's text
	Text *string `json:"text" yaml:"text"`
	// The identity's hand size
	HandSize *int `json:"hand_size,omitempty" yaml:"hand_size,omitempty"`
	// The character's hit points (usually an Ally, Identity, or Minion)
	HitPoints *int `json:"hit_points,omitempty" yaml:"hit_points,omitempty"`
	// The number of hit points per player, usually for a Villain, but possibly a Minion
	HitPointsPerPlayer *int `json:"hit_points_per_player,omitempty" yaml:"hit_points_per_player,omitempty"`
	// The villain or main scheme's stage - this is a string because some villains or schemes may be 1A, 1B, etc.
	Stage *string `json:"stage,omitempty" yaml:"stage,omitempty"`
	// The encounter card's boost icons
	BoostIcons *int `json:"boost_icons,omitempty" yaml:"boost_icons,omitempty"`
	// The text associated with a star icon on the encounter card
	StarText *string `json:"star_text,omitempty" yaml:"star_text,omitempty"`
	// Acceleration, Amplify, Crisis, Hazard
	// This is called encounter icons and not side scheme icons because some minions have the Amplify icon
	EncounterIcons []string `json:"encounter_icons,omitempty" yaml:"encounter_icons,omitempty"`
	// The main scheme or side scheme's fixed starting threat, which may be further modified by per-player threat
	StartingThreat *int `json:"starting_threat,omitempty" yaml:"starting_threat,omitempty"`
	// The main scheme or side scheme's starting threat per player
	StartingThreatPerPlayer *int `json:"starting_threat_per_player,omitempty" yaml:"starting_threat_per_player,omitempty"`
	// The main scheme's (or in the future, side scheme's?) acceleration threat
	// In the case of schemes with variable threat (such as Mutagen Cloud), we should use -1
	AccelerationThreat *int `json:"acceleration_threat,omitempty" yaml:"acceleration_threat,omitempty"`
	// The main scheme's (or in the future, side scheme's?) acceleration threat per player
	AccelerationThreatPerPlayer *int `json:"acceleration_threat_per_player,omitempty" yaml:"acceleration_threat_per_player,omitempty"`
	// The amount of threat required to complete the scheme - currently, I believe this is always per player
	TargetThreat *int `json:"target_threat,omitempty" yaml:"completion_threat,omitempty"`
	// The amount of threat per player required to complete the scheme
	TargetThreatPerPlayer *int `json:"target_threat_per_player,omitempty" yaml:"completion_threat_per_player,omitempty"`
	// The flavor text of the card
	FlavorText *string `json:"flavor_text,omitempty" yaml:"flavor_text,omitempty"`
	// The resources generated by the card
	*Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	// The HTTP/HTTPS URL where the card image can be downloaded
	ImageURL *string `json:"image_url" yaml:"image_url"`
	// The MarvelCDB.com URL for the card
	MarvelCDBURL *string `json:"marvelcdb_url" yaml:"marvelcdb_url"`
	// The card's illustrator
	Illustrator *string `json:"illustrator,omitempty" yaml:"illustrator,omitempty"`
}

// The amount of resources generated by a card when paying a cost
type Resources struct {
	Energy   *int `json:"energy,omitempty" yaml:"energy,omitempty"`
	Mental   *int `json:"mental,omitempty" yaml:"mental,omitempty"`
	Physical *int `json:"physical,omitempty" yaml:"physical,omitempty"`
	Wild     *int `json:"wild,omitempty" yaml:"wild,omitempty"`
}

// Pack is a pack of Marvel Champions cards, such as the Core set or the Rise of Red Skull expansion, or a hero or
// villain pack such as Captain America or Green Goblin.
type Pack struct {
	Name     string `json:"name" yaml:"name"`         // For PnP sets, we will create our own name.
	SKU      string `json:"sku" yaml:"sku"`           // For PnP sets, we will create our own SKU.
	Position *int   `json:"position" yaml:"position"` // Position may not exist for entities like status cards.
	Quantity *int   `json:"quantity" yaml:"quantity"` // Quantity may not exist for entities like status cards.
}

// Set is a subset of Marvel Champions cards, such as the Spider-Man Hero deck or the Expert encounter cards
type Set struct {
	Name string `json:"name" yaml:"name"`
}

// DownloadImages will attempt to download all images for the card from S3 to local storage.
func (c *Card) DownloadImages() (err error) {
	if len(c.Faces) == 0 {
		return fmt.Errorf("unable to download images for %s: no faces\n", c.Names[0])
	}
	// We want to save cards to images/<SKU>/<image_name>
	// We will use the first SKU for the card as the image path
	for _, face := range c.Faces {
		// No image to save - try the next face
		if face.ImageURL == nil {
			return fmt.Errorf("no images for %s", face.Name)
		}
		if len(c.Packs) == 0 {
			return fmt.Errorf("unable to determine pack for %s", face.Name)
		}
		// Determine where to save the image to
		imageSlice := strings.Split(*face.ImageURL, "/")
		imageName := imageSlice[len(imageSlice)-1]
		imageDir := fmt.Sprintf("%s/%s", IMAGE_BASEDIR, c.Packs[0].SKU)
		imagePath := fmt.Sprintf("%s/%s", imageDir, imageName)
		// Open the image file, or create it if it doesn't exist
		if _, err := os.Stat(imageDir); os.IsNotExist(err) {
			os.Mkdir(imageDir, 0755)
		}
		f, err := os.OpenFile(imagePath, os.O_RDWR|os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("unable to open file path: %s", imagePath)
		}
		// We have to use the Stat() method to have access to the file size
		fi, err := f.Stat()
		if err != nil {
			return fmt.Errorf("error reading file metadata: %s\n", imagePath)
		}
		// If the file is larger than 0 bytes, we will not attempt to download it
		if fi.Size() > 0 {
			continue
		}
		// At this point, we attempt to make an HTTP call to download the image and save it locally
		resp, err := http.Get(*face.ImageURL)
		if err != nil {
			return fmt.Errorf("error retrieving image from %s: %v\n", face.ImageURL, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("bad http status code retrieving image from %s: %d\n", *face.ImageURL, resp.StatusCode)
		}
		// Decode the image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return fmt.Errorf("error decoding image from %s: %w\n", face.ImageURL, err)
		}
		// Write the image to our file
		err = png.Encode(f, img)
		if err != nil {
			return fmt.Errorf("error writing image: %w\n", err)
		}
	}
	return nil
}

// NameMatch searches for an exact match between the query string and one of the
// Card's Name strings.
func (c *Card) NameMatch(query string) bool {
	for _, name := range c.Names {
		if strings.EqualFold(query, name) {
			return true
		}
	}
	return false
}

// IsScheme returns whether the card face is a scheme or not.
func (f *Face) IsScheme() bool {
	if strings.ToLower(f.Type) == "main scheme" || strings.ToLower(f.Type) == "side scheme" {
		return true
	}
	return false
}
