package card

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const IMAGE_BASEDIR = "images"

// Card is a physical representation of a Marvel Champions card. Each instance of Card should represent the same
// physical copy of the card. For example, Peter Parker/Spider-Man is a single card. However, Rhino I, Rhino II, and
// Rhino III is three separate cards.
type Card struct {
	Names      []string `json:"names" yaml:"names"`                               // Names in which the card is known by.
	Packs      []*Pack  `json:"packs" yaml:"packs"`                               // Packs in which the card has appeared.
	Sets       []*Set   `json:"sets"  yaml:"sets"`                                // Sets in which the card is a member.
	Faces      []*Face  `json:"faces" yaml:"faces"`                               // Supports multiple sides of the card.
	Horizontal bool     `json:"horizontal,omitempty" yaml:"horizontal,omitempty"` // Whether the card is rotated horizontally.
}

// MarvelCDB card is a card pulled from the MarvelCDB API.
type MarvelCDBCard struct {
	PackCode              string  `json:"pack_code"`
	PackName              string  `json:"pack_name"`
	TypeCode              string  `json:"type_code"`
	TypeName              string  `json:"type_name"`
	FactionCode           string  `json:"faction_code"`
	FactionName           string  `json:"faction_name"`
	SetCode               string  `json:"card_set_code"`
	SetName               string  `json:"card_set_name"`
	Position              int     `json:"position"`
	Code                  string  `json:"code"`
	Name                  string  `json:"name"`
	RealName              string  `json:"real_name"`
	Subname               *string `json:"subname"`
	Cost                  *int    `json:"cost"`
	Text                  *string `json:"text"`
	RealText              *string `json:"real_text"`
	Quantity              int     `json:"quantity"`
	ResourceEnergy        *int    `json:"resource_energy"`
	ResourceMental        *int    `json:"resource_mental"`
	ResourcePhysical      *int    `json:"resource_physical"`
	ResourceWild          *int    `json:"resource_wild"`
	HealthPerHero         bool    `json:"health_per_hero"`
	Thwart                *int    `json:"thwart"`
	ThwartStar            *bool
	ThwartText            *string `json:"thwart_text"`
	ThwartCost            *int    `json:"thwart_cost"`
	Scheme                *int    `json:"scheme"`
	SchemeText            *string `json:"scheme_text"`
	Attack                *int    `json:"attack"`
	AttackText            *string `json:"attack_text"`
	AttackCost            *int    `json:"attack_cost"`
	Defense               *int    `json:"defense"`
	DefenseText           *string `json:"defense_text"`
	DefenseCost           *int    `json:"defense_cost"`
	Recover               *int    `json:"recover"`
	BaseThreat            *int    `json:"base_threat"`
	BaseThreatFixed       *bool   `json:"base_threat_fixed"`
	EscalationThreat      *int    `json:"escalation_threat"`
	EscalationThreatFixed *bool   `json:"escalation_threat_fixed"`
	Threat                *int    `json:"threat"`
	ThreatFixed           *bool   `json:"threat_fixed"`
	DeckLimit             int     `json:"deck_limit"`
	HandSize              *int    `json:hand_size`
	Traits                *string `json:"traits"`
	RealTraits            string  `json:"real_traits"`
	Flavor                *string `json:"flavor"`
	Boost                 *int    `json:"boost"`
	BoostText             *string `json:"boost_text"`
	Health                *int    `json:"health"`
	IsUnique              bool    `json:"is_unique"`
	Hidden                bool    `json:"hidden"`
	DoubleSided           bool    `json:"double_sided"`
	BackText              *string `json:"back_text"`
	BackFlavor            *string `json:"back_flavor"`
	OCTGNId               string  `json:"octgn_id"`
	URL                   *string `json:"url"`
	ImageSrc              string  `json:"imagesrc"`
	Spoiler               int     `json:"spoiler"`
	BackImageSrc          string  `json:"backimagesrc"`
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

// IsScheme returns whether the card face is a scheme or not.
func (f *Face) IsScheme() bool {
	if strings.ToLower(f.Type) == "main scheme" || strings.ToLower(f.Type) == "side scheme" {
		return true
	}
	return false
}

// Convert converts a MarvelCDBCard into our Card.
func (mcdb *MarvelCDBCard) Convert() *Card {
	card := &Card{}
	card.Names = append(card.Names, mcdb.Name)
	pack := &Pack{}
	// Rotate schemes and side schemes
	if mcdb.TypeCode == "main_scheme" || mcdb.TypeCode == "side_scheme" {
		card.Horizontal = true
	}
	switch mcdb.PackName {
	case "Core Set":
		pack.Name = "Core Set"
		pack.SKU = "MC01en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "The Green Goblin":
		pack.Name = "The Green Goblin"
		pack.SKU = "MC02en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "The Wrecking Crew":
		pack.Name = "The Wrecking Crew"
		pack.SKU = "MC03en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Captain America":
		pack.Name = "Captain America"
		pack.SKU = "MC04en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Ms. Marvel":
		pack.Name = "Ms. Marvel"
		pack.SKU = "MC05en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Thor":
		pack.Name = "Thor"
		pack.SKU = "MC06en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Black Widow":
		pack.Name = "Black Widow"
		pack.SKU = "MC07en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Doctor Strange":
		pack.Name = "Dr. Strange"
		pack.SKU = "MC08en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Hulk":
		pack.Name = "Hulk"
		pack.SKU = "MC09en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "The Rise of Red Skull":
		pack.Name = "The Rise of Red Skull"
		pack.SKU = "MC10en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "The Once and Future Kang":
		pack.Name = "The Once and Future Kang"
		pack.SKU = "MC11en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Ant-man":
		pack.Name = "Ant-Man"
		pack.SKU = "MC12en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Wasp":
		pack.Name = "Wasp"
		pack.SKU = "MC13en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Quicksilver":
		pack.Name = "Quicksilver"
		pack.SKU = "MC14en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Scarlet Witch":
		pack.Name = "Scarlet Witch"
		pack.SKU = "MC15en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Galaxy's Most Wanted":
		pack.Name = "Galaxy’s Most Wanted"
		pack.SKU = "MC16en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Star-Lord":
		pack.Name = "Star-Lord"
		pack.SKU = "MC17en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Gamora":
		pack.Name = "Gamora"
		pack.SKU = "MC18en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Drax":
		pack.Name = "Drax"
		pack.SKU = "MC19en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	case "Venom":
		pack.Name = "Venom"
		pack.SKU = "MC20en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	// PnP cards
	case "Ronan Modular Set":
		pack.Name = "Ronan Modular Set"
		pack.SKU = "PNP01en"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	default:
		pack.Name = mcdb.PackName
		pack.SKU = "Unknown"
		pack.Position = &mcdb.Position
		pack.Quantity = &mcdb.Quantity
	}

	card.Packs = append(card.Packs, pack)

	face := &Face{}
	// Name
	face.Name = mcdb.Name
	// Subtitle
	face.Subtitle = mcdb.Subname
	// Cost
	if mcdb.Cost != nil {
		face.Cost = mcdb.Cost
	}
	// Unique
	face.Unique = mcdb.IsUnique
	// Type
	face.Type = func(s string) string {
		return strings.Title(strings.ReplaceAll(s, "_", " "))
	}(mcdb.TypeCode)
	// Aspect
	switch mcdb.FactionName {
	case "Aggression", "Basic", "Justice", "Leadership", "Protection":
		face.Aspect = append(face.Aspect, mcdb.FactionName)
	}
	// RecoverValue
	face.RecoverValue = mcdb.Recover
	// SchemeValue
	face.SchemeValue = mcdb.Scheme
	// SchemeText
	face.SchemeText = mcdb.SchemeText
	// ThwartValue
	face.ThwartValue = mcdb.Thwart
	// ThwartText
	face.ThwartText = mcdb.ThwartText
	// ThwartConsequential
	face.ThwartConsequential = mcdb.ThwartCost
	// AttackValue
	face.AttackValue = mcdb.Attack
	// AttackText
	face.AttackText = mcdb.AttackText
	// AttackConsequential
	face.AttackConsequential = mcdb.AttackCost
	// DefenseValue
	face.DefenseValue = mcdb.Defense
	// DefenseText
	face.DefenseText = mcdb.DefenseText
	// Traits
	if mcdb.Traits != nil {
		face.Traits = func(s string) []string {
			var traits = []string{}
			// We can't use strings.Fields because of traits like "Accuser Corps"
			rawTraits := strings.SplitAfter(s, ". ")
			// Trim trailing period if it is not an acronym like "S.H.I.E.L.D."
			for _, trait := range rawTraits {
				trait = strings.TrimSpace(trait)
				periodSearch := regexp.MustCompile(`\.`)
				matches := periodSearch.FindAllStringIndex(trait, -1)
				if len(matches) == 1 {
					trait = strings.TrimRight(trait, ".")
					traits = append(traits, trait)
				}
			}
			return traits
		}(*mcdb.Traits)
	}
	// Keywords
	// Keywords are tricky, because MarvelCDB doesn't actually track them, but we do.
	keywords := []string{}
	if mcdb.Text != nil {
		for _, keyword := range mcdbKeywords {
			if strings.Contains(*mcdb.Text, keyword) {
				keyword = strings.TrimRight(keyword, ".")
				if keyword == "Team Up" {
					keyword = "Team-Up"
				}
				keywords = append(keywords, keyword)
			}
		}
	}
	if len(keywords) > 0 {
		face.Keywords = keywords
	}
	// Text
	if mcdb.Text != nil {
		face.Text = func(s string) *string {
			s = strings.ReplaceAll(s, "<b>", "**")
			s = strings.ReplaceAll(s, "</b>", "**")
			s = strings.ReplaceAll(s, "<i>", "_")
			s = strings.ReplaceAll(s, "</i>", "_")
			return &s
		}(*mcdb.Text)
	}
	// Hand Size
	face.HandSize = mcdb.HandSize
	// Hit Points
	if mcdb.Health != nil && mcdb.HealthPerHero == false {
		face.HitPoints = mcdb.Health
	}
	// Hit Points Per Player
	if mcdb.Health != nil && mcdb.HealthPerHero == true {
		face.HitPointsPerPlayer = mcdb.Health
	}
	// Starting Threat
	if mcdb.BaseThreat != nil && *mcdb.BaseThreatFixed == true {
		face.StartingThreat = mcdb.BaseThreat
	}
	// Starting Threat Per Player
	if mcdb.BaseThreat != nil && *mcdb.BaseThreatFixed == false {
		face.StartingThreatPerPlayer = mcdb.BaseThreat
	}
	// Acceleration Threat
	if mcdb.EscalationThreat != nil && *mcdb.EscalationThreatFixed == true {
		face.AccelerationThreat = mcdb.EscalationThreat
	}
	// Acceleration Threat Per Player
	if mcdb.EscalationThreat != nil && *mcdb.EscalationThreatFixed == false {
		face.AccelerationThreatPerPlayer = mcdb.EscalationThreat
	}
	// Target Threat
	if mcdb.Threat != nil && *mcdb.ThreatFixed == true {
		face.TargetThreat = mcdb.Threat
	}
	// Target Threat Per Player
	if mcdb.Threat != nil && *mcdb.ThreatFixed == false {
		face.TargetThreatPerPlayer = mcdb.Threat
	}
	// BoostIcons
	if mcdb.Boost != nil {
		face.BoostIcons = mcdb.Boost
	} else {
		// MCDB doesn't differentiate between a card that can't have boost icons, and a card with 0 boost icons - we do
		if mcdb.FactionCode == "encounter" {
			switch mcdb.TypeCode {
			case "attachment", "minion", "obligation", "side_scheme", "treachery":
				boostIcons := 0
				face.BoostIcons = &boostIcons
			}
		}
	}
	// StarBoost
	if mcdb.BoostText != nil {
		face.StarText = mcdb.BoostText
	}
	// Flavor Text
	if mcdb.Flavor != nil {
		face.FlavorText = mcdb.Flavor
	}
	// Resources
	resources := &Resources{}
	if mcdb.ResourceEnergy != nil || mcdb.ResourceMental != nil || mcdb.ResourcePhysical != nil || mcdb.ResourceWild != nil {
		resources.Energy = mcdb.ResourceEnergy
		resources.Mental = mcdb.ResourceMental
		resources.Physical = mcdb.ResourcePhysical
		resources.Wild = mcdb.ResourceWild
		face.Resources = resources
	}
	// Image URL
	if pack.SKU != "Unknown" {
		lowerSKU := strings.ToLower(card.Packs[0].SKU)
		// If the last character of the MCDB URL is a letter, we'll append that
		var imageURL string
		if mcdb.URL != nil {
			mcdbURL := *mcdb.URL
			lastCharacter := mcdbURL[len(mcdbURL)-1:]
			if unicode.IsLetter([]rune(lastCharacter)[0]) {
				imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%d%s.png", lowerSKU, mcdb.Position, strings.ToUpper(lastCharacter))
			} else {
				// If it's a main scheme, this is the B side
				if mcdb.TypeCode != "main_scheme" {
					imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%d.png", lowerSKU, mcdb.Position)
				} else {
					imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%dB.png", lowerSKU, mcdb.Position)
				}
			}
		}
		face.ImageURL = &imageURL
	}
	// MarvelCDB.com URL
	if mcdb.URL != nil {
		mcdbUrl := strings.ReplaceAll(*mcdb.URL, "\\", "")
		face.MarvelCDBURL = &mcdbUrl
	}
	card.Faces = append(card.Faces, face)

	// If the card is a main scheme, we want both sides
	if mcdb.DoubleSided == true && mcdb.TypeCode == "main_scheme" {
		backface := &Face{}
		// Name
		backface.Name = mcdb.Name
		// Subtitle
		backface.Subtitle = mcdb.Subname
		// Cost
		if mcdb.Cost != nil {
			backface.Cost = mcdb.Cost
		}
		// Unique
		backface.Unique = mcdb.IsUnique
		// Type
		backface.Type = func(s string) string {
			return strings.Title(strings.ReplaceAll(s, "_", " "))
		}(mcdb.TypeCode)
		// Aspect
		switch mcdb.FactionName {
		case "Aggression", "Basic", "Justice", "Leadership", "Protection":
			backface.Aspect = append(face.Aspect, mcdb.FactionName)
		}
		// RecoverValue
		backface.RecoverValue = mcdb.Recover
		// SchemeValue
		backface.SchemeValue = mcdb.Scheme
		// SchemeText
		backface.SchemeText = mcdb.SchemeText
		// ThwartValue
		backface.ThwartValue = mcdb.Thwart
		// ThwartText
		backface.ThwartText = mcdb.ThwartText
		// ThwartConsequential
		backface.ThwartConsequential = mcdb.ThwartCost
		// AttackValue
		backface.AttackValue = mcdb.Attack
		// AttackText
		backface.AttackText = mcdb.AttackText
		// AttackConsequential
		backface.AttackConsequential = mcdb.AttackCost
		// DefenseValue
		backface.DefenseValue = mcdb.Defense
		// DefenseText
		backface.DefenseText = mcdb.DefenseText
		// Traits
		if mcdb.Traits != nil {
			backface.Traits = func(s string) []string {
				var traits = []string{}
				// We can't use strings.Fields because of traits like "Accuser Corps"
				rawTraits := strings.SplitAfter(s, ". ")
				// Trim trailing period if it is not an acronym like "S.H.I.E.L.D."
				for _, trait := range rawTraits {
					trait = strings.TrimSpace(trait)
					periodSearch := regexp.MustCompile(`\.`)
					matches := periodSearch.FindAllStringIndex(trait, -1)
					if len(matches) == 1 {
						trait = strings.TrimRight(trait, ".")
						traits = append(traits, trait)
					}
				}
				return traits
			}(*mcdb.Traits)
		}
		// Keywords
		// Keywords are tricky, because MarvelCDB doesn't actually track them, but we do.
		keywords := []string{}
		if mcdb.Text != nil {
			for _, keyword := range mcdbKeywords {
				if strings.Contains(*mcdb.Text, keyword) {
					keyword = strings.TrimRight(keyword, ".")
					if keyword == "Team Up" {
						keyword = "Team-Up"
					}
					keywords = append(keywords, keyword)
				}
			}
		}
		if len(keywords) > 0 {
			backface.Keywords = keywords
		}
		// Text
		if mcdb.BackText != nil {
			backface.Text = func(s string) *string {
				s = strings.ReplaceAll(s, "<b>", "**")
				s = strings.ReplaceAll(s, "</b>", "**")
				s = strings.ReplaceAll(s, "<i>", "_")
				s = strings.ReplaceAll(s, "</i>", "_")
				return &s
			}(*mcdb.BackText)
		}
		// Hand Size
		backface.HandSize = mcdb.HandSize
		// Hit Points
		if mcdb.Health != nil && mcdb.HealthPerHero == false {
			backface.HitPoints = mcdb.Health
		}
		// Hit Points Per Player
		if mcdb.Health != nil && mcdb.HealthPerHero == true {
			backface.HitPointsPerPlayer = mcdb.Health
		}
		// Starting Threat
		if mcdb.BaseThreat != nil && *mcdb.BaseThreatFixed == true {
			backface.StartingThreat = mcdb.BaseThreat
		}
		// Starting Threat Per Player
		if mcdb.BaseThreat != nil && *mcdb.BaseThreatFixed == false {
			backface.StartingThreatPerPlayer = mcdb.BaseThreat
		}
		// Acceleration Threat
		if mcdb.EscalationThreat != nil && *mcdb.EscalationThreatFixed == true {
			backface.AccelerationThreat = mcdb.EscalationThreat
		}
		// Acceleration Threat Per Player
		if mcdb.EscalationThreat != nil && *mcdb.EscalationThreatFixed == false {
			backface.AccelerationThreatPerPlayer = mcdb.EscalationThreat
		}
		// Target Threat
		if mcdb.Threat != nil && *mcdb.ThreatFixed == true {
			backface.TargetThreat = mcdb.Threat
		}
		// Target Threat Per Player
		if mcdb.Threat != nil && *mcdb.ThreatFixed == false {
			backface.TargetThreatPerPlayer = mcdb.Threat
		}
		// BoostIcons
		if mcdb.Boost != nil {
			backface.BoostIcons = mcdb.Boost
		} else {
			// MCDB doesn't differentiate between a card that can't have boost icons, and a card with 0 boost icons - we do
			if mcdb.FactionCode == "encounter" {
				switch mcdb.TypeCode {
				case "attachment", "minion", "obligation", "side_scheme", "treachery":
					boostIcons := 0
					backface.BoostIcons = &boostIcons
				}
			}
		}
		// StarBoost
		if mcdb.BoostText != nil {
			backface.StarText = mcdb.BoostText
		}
		// Flavor Text
		if mcdb.BackFlavor != nil {
			backface.FlavorText = mcdb.BackFlavor
		}
		// Resources
		resources := &Resources{}
		if mcdb.ResourceEnergy != nil || mcdb.ResourceMental != nil || mcdb.ResourcePhysical != nil || mcdb.ResourceWild != nil {
			resources.Energy = mcdb.ResourceEnergy
			resources.Mental = mcdb.ResourceMental
			resources.Physical = mcdb.ResourcePhysical
			resources.Wild = mcdb.ResourceWild
			backface.Resources = resources
		}
		// Image URL
		if pack.SKU != "Unknown" {
			lowerSKU := strings.ToLower(card.Packs[0].SKU)
			// If the last character of the MCDB URL is a letter, we'll append that
			var imageURL string
			if mcdb.URL != nil {
				mcdbURL := *mcdb.URL
				lastCharacter := mcdbURL[len(mcdbURL)-1:]
				if unicode.IsLetter([]rune(lastCharacter)[0]) {
					imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%d%s.png", lowerSKU, mcdb.Position, strings.ToUpper(lastCharacter))
				} else {
					// If it's a main scheme, this is the B side
					if mcdb.TypeCode != "main_scheme" {
						imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%d.png", lowerSKU, mcdb.Position)
					} else {
						imageURL = fmt.Sprintf("https://marvel-champions-cards.s3.us-west-2.amazonaws.com/%s/%dA.png", lowerSKU, mcdb.Position)
					}
				}
			}
			backface.ImageURL = &imageURL
		}
		// MarvelCDB.com URL
		if mcdb.URL != nil {
			mcdbUrl := strings.ReplaceAll(*mcdb.URL, "\\", "")
			backface.MarvelCDBURL = &mcdbUrl
		}
		card.Faces = append(card.Faces, backface)
	}

	/*
		yamlText, err := yaml.Marshal(card)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(yamlText))
		}
	*/

	return card
}
