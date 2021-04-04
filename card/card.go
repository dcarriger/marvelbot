package card

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Card holds a representation of a card from the MarvelCDB API.
type Card struct {
	PackCode              string `json:"pack_code"`
	PackName              string `json:"pack_name"`
	TypeCode              string `json:"type_code"`
	TypeName              string `json:"type_name"`
	FactionCode           string `json:"faction_code"`
	FactionName           string `json:"faction_name"`
	SetCode               string `json:"card_set_code"`
	SetName               string `json:"card_set_name"`
	Position              int    `json:"position"`
	Code                  string `json:"code"`
	Name                  string `json:"name"`
	RealName              string `json:"real_name"`
	Subname               string `json:"subname"`
	Cost                  int    `json:"cost"`
	Text                  string `json:"text"`
	RealText              string `json:"real_text"`
	Quantity              int    `json:"quantity"`
	ResourceEnergy        int    `json:"resource_energy"`
	ResourceMental        int    `json:"resource_mental"`
	ResourcePhysical      int    `json:"resource_physical"`
	ResourceWild          int    `json:"resource_wild"`
	HealthPerHero         bool   `json:"health_per_hero"`
	Thwart                int    `json:"thwart"`
	ThwartCost            int    `json:"thwart_cost"`
	Attack                int    `json:"attack"`
	AttackCost            int    `json:"attack_cost"`
	BaseThreatFixed       bool   `json:"base_threat_fixed"`
	EscalationThreatFixed bool   `json:"escalation_threat_fixed"`
	ThreatFixed           bool   `json:"threat_fixed"`
	DeckLimit             int    `json:"deck_limit"`
	Traits                string `json:"traits"`
	RealTraits            string `json:"real_traits"`
	Flavor                string `json:"flavor"`
	IsUnique              bool   `json:"is_unique"`
	Hidden                bool   `json:"hidden"`
	DoubleSided           bool   `json:"double_sided"`
	BackText              string `json:"back_text"`
	BackFlavor            string `json:"back_flavor"`
	OCTGNId               string `json:"octgn_id"`
	URL                   string `json:"url"`
	ImageSrc              string `json:"imagesrc"`
	Spoiler               int    `json:"spoiler"`
	BackImageSrc          string `json:"backimagesrc"`
}

// Cards is simply a slice of Card.
type Cards []*Card

// CostIcon returns the appropriate Discord emoji string based on the card cost.
func (c Card) CostIcon() string {
	switch c.Cost {
	case 0:
		// Resources can only be used to pay costs and do not have a cost themselves.
		if c.TypeCode == "resource" {
			return Resource
		}
		return Zero
	case 1:
		return One
	case 2:
		return Two
	case 3:
		return Three
	case 4:
		return Four
	case 5:
		return Five
	case 6:
		return Six
	default:
		return ""
	}
}

// EmbedString is a string output of a Card for use in a Discord embed field.
func (c Card) EmbedString() string {
	return fmt.Sprintf("[%s](%s)", c.Name, c.URL)
}

// Normalize returns a normalized string of the Card name.
func (c Card) Normalize() (cardName string) {
	// Lowercase the card name
	cardName = strings.ToLower(c.Name)
	// Remove punctuation so we don't have to deal with "Wakanda Forever!" or "Get Behind Me!" et al
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	cardName = re.ReplaceAllString(cardName, "")
	return cardName
}

// GetImagePath will get the path to the local card images.
func (c Card) GetImagePath() (path string, err error) {
	// If the card doesn't have an image, there's nothing to do.
	if c.ImageSrc == "" {
		return "", fmt.Errorf("No image found for: %v\n", c.Name)
	}

	// Split the image URL into parts
	imageSlice := strings.Split(c.ImageSrc, "/")
	imageName := strings.ToLower(imageSlice[len(imageSlice)-1])

	// We'll also strip the extension from the file name
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName))

	// We want to save images as PNG so we'll define our image path here
	imagePath := "images/" + imageName + ".png"

	return imagePath, nil
}

// downloadImage takes an image URL and downloads a single image.
func downloadImage(path string, scheme bool) (err error) {
	// Convert the URL to a file path
	imagePath := GetImagePath(path, false)
	// Open the file, or create it if it doesn't exist
	f, err := os.OpenFile(imagePath, os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("unable to open path: %v\n", imagePath)
	}
	// We have to use the Stat() method to have access to the file size
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Error accessing metadata: %v\n", imagePath)
	}
	// If the size is 0, we need to get the image from MarvelCDB
	if fi.Size() == 0 {
		// Make a GET request for the image and close the connection
		// MarvelCDB has bad images for some cards
		var getURL string
		// TODO - cards with a bad path?
		getURL = "https://marvelcdb.com" + path
		resp, err := http.Get(getURL)
		if err != nil {
			return fmt.Errorf("Error accessing image from MarvelCDB: %w\n", err)
		}
		defer resp.Body.Close()
		// Decode the image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return fmt.Errorf("Error decoding image: %w\n", err)
		}
		// Resize the image to 300x419 (vertical) or 419x300 (horizontal)
		if scheme == false {
			img = imaging.Resize(img, 300, 419, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 419, 300, imaging.Lanczos)
		}
		// Convert the image to PNG and write it to a file
		err = png.Encode(f, img)
		if err != nil {
			return fmt.Errorf("Error writing encoded image: %w\n", err)
		}
	}
	// Now check for a rotated file
	if scheme == true {
		rotatedPath := GetImagePath(path, true)
		rf, err := os.OpenFile(rotatedPath, os.O_RDWR|os.O_CREATE, 0644)
		defer rf.Close()
		if err != nil {
			return fmt.Errorf("unable to open path: %v\n", rotatedPath)
		}
		rfi, err := rf.Stat()
		if err != nil {
			return fmt.Errorf("error accessing metadata: %v\n", rotatedPath)
		}
		if rfi.Size() == 0 {
			// Rotate the image and save it
			err = RotateImage(f, rotatedPath)
			if err != nil {
				return fmt.Errorf("error rotating file: %w\n", err)
			}
		}
	}
	return nil
}

// GetImages is a function to return all images associated with the Card.
func (c Card) GetImages() (images []string) {
	if c.BackImageSrc != "" {
		images = append(images, c.BackImageSrc)
	}
	if c.ImageSrc != "" {
		images = append(images, c.ImageSrc)
	}
	return images
}

// IsScheme returns whether the card is a scheme or not.
func (c Card) IsScheme() bool {
	if c.TypeCode == "main_scheme" || c.TypeCode == "side_scheme" {
		return true
	}
	return false
}

// DownloadImages will download the card images from Marvel CDB.
func (c Card) DownloadImages() (err error) {
	if c.ImageSrc == "" {
		return fmt.Errorf("No image found for: %v\n", c.Name)
	}
	// We check if it's a scheme so we know whether we need to save a rotated image as well
	scheme := c.IsScheme()
	err = downloadImage(c.ImageSrc, scheme)
	if err != nil {
		return err
	}
	// If the card is two-sided, we need the reverse image as well
	if c.BackImageSrc != "" {
		err := downloadImage(c.BackImageSrc, scheme)
		if err != nil {
			return err
		}
	}
	return nil

	/*
		// Split the image URL into parts
		imageSlice := strings.Split(c.ImageSrc, "/")
		imageName := strings.ToLower(imageSlice[len(imageSlice)-1])

		// We'll also strip the extension from the file name
		imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName))

		// We want to save images as PNG so we'll define our image path here
		imagePath := "images/" + imageName + ".png"

		// We'll want a rotated version of some images as well
		imageRotatedPath := "images/" + imageName + "_rotated.png"
		_ = imageRotatedPath

		// Now comes the logic to download the images if we don't have them already
		// 0666 used because default umask will modify to 0644
		f, err := os.OpenFile(imagePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("Unable to open path: %v\n", imagePath)
		}

		// We have to use the Stat() method to have access to the file size
		fi, err := f.Stat()
		if err != nil {
			return fmt.Errorf("Error accessing metadata: %v\n", imagePath)
		}

		// If the size is 0, we need to get the image from MarvelCDB
		if fi.Size() == 0 {
			// Make a GET request for the image and close the connection
			// MarvelCDB has bad images for some cards
			var getURL string
			switch c.Name {
			case "Spider-Man":
				getURL = "https://lcgcdn.s3.amazonaws.com/mc/MC01en_1A.jpg"
			case "Peter Parker":
				getURL = "https://lcgcdn.s3.amazonaws.com/mc/MC01en_1B.jpg"
			case "She-Hulk":
				getURL = "https://lcgcdn.s3.amazonaws.com/mc/MC01en_19A.jpg"
			case "Jennifer Walters":
				getURL = "https://lcgcdn.s3.amazonaws.com/mc/MC01en_19B.jpg"
			case "Generation Why?":
				getURL = "https://lcgcdn.s3.amazonaws.com/mc/MC05en_26.jpg"
			default:
				getURL = "https://marvelcdb.com" + c.ImageSrc
			}

			resp, err := http.Get(getURL)
			if err != nil {
				return fmt.Errorf("Error accessing image from MarvelCDB: %w\n", err)
			}
			defer resp.Body.Close()

			// Decode the image
			img, _, err := image.Decode(resp.Body)
			if err != nil {
				return fmt.Errorf("Error decoding image: %w\n", err)
			}

			// Resize the image to 300x419 (vertical) or 419x300 (horizontal)
			if c.TypeCode != "main_scheme" && c.TypeCode != "side_scheme" {
				img = imaging.Resize(img, 300, 419, imaging.Lanczos)
			} else {
				img = imaging.Resize(img, 419, 300, imaging.Lanczos)
			}

			// Convert the image to PNG and write it to a file
			err = png.Encode(f, img)
			if err != nil {
				return fmt.Errorf("Error writing encoded image: %w\n", err)
			}
		}

		return nil
	*/
}

// SortSlice is a Cards method to sort by Cost,Name.
func (c Cards) SortSlice() Cards {
	sort.Slice(c, func(i, j int) bool {
		// If we have a resource, that should be sorted to the top
		if c[i].TypeCode == "resource" && c[j].TypeCode != "resource" {
			return true
		} else if c[i].TypeCode != "resource" && c[j].TypeCode == "resource" {
			return false
		}
		// If the Cost is lower, that should also be sorted to the top. If the Cost is equal, we'll proceed to a Name
		// comparison.
		if c[i].Cost < c[j].Cost {
			return true
		} else if c[i].Cost > c[j].Cost {
			return false
		}
		// Sort alphabetically by name, Cost being equal.
		if c[i].Name < c[j].Name {
			return true
		}
		return false
	})
	return c
}

// RotateImage takes an image file and saves a rotated copy of it.
func RotateImage(r io.Reader, path string) (err error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	rotatedImg := imaging.Rotate90(img)
	err = imaging.Save(rotatedImg, path)
	if err != nil {
		return err
	}
	return nil
}

// GetImagePath takes a URL and converts it into a local file path.
func GetImagePath(path string, rotate bool) string {
	if path == "" {
		return ""
	}
	// Convert the URL to a usable path
	imageSlice := strings.Split(path, "/")
	imageName := strings.ToLower(imageSlice[len(imageSlice)-1])
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName))
	// We want to save images as PNG so we'll define our image path here
	var imagePath string
	if rotate == true {
		imagePath = "images/" + imageName + "_rotated.png"
	} else {
		imagePath = "images/" + imageName + ".png"
	}
	return imagePath
}
