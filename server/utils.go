package server

import (
	"fmt"
	gim "github.com/ozankasikci/go-image-merge"
	"github.com/segmentio/ksuid"
	"image/png"
	"marvelbot/card"
	"math"
	"os"
	"strings"
)

const Director = "243490403800711169"

// buildImage takes a slice of cards and builds a single image from them. It is assumed that some pre-processing is
// done on the cards first to separate horizontal from vertical cards.
func buildImage(cards []*card.Card) (fileName string, cardsWithErrors []*card.Card, err error) {
	// Slice to hold all of our grids
	grids := []*gim.Grid{}
	// Now, we'll add all of the cards to our grid
	for _, c := range cards {
		err := c.DownloadImages()
		if err != nil {
			cardsWithErrors = append(cardsWithErrors, c)
			continue
		}
		for _, cardFace := range c.Faces {
			imageSlice := strings.Split(*cardFace.ImageURL, "/")
			imageName := imageSlice[len(imageSlice)-1]
			imageDir := fmt.Sprintf("%s/%s", IMAGE_BASEDIR, c.Packs[0].SKU)
			imagePath := fmt.Sprintf("%s/%s", imageDir, imageName)
			// Add cards to their respective Grid based on orientation (vertical or horizontal)
			grid := &gim.Grid{
				ImageFilePath: imagePath,
			}
			grids = append(grids, grid)
		}
	}
	// If the grids slice is empty, we are unable to return any images
	if len(grids) == 0 {
		return "", cardsWithErrors, fmt.Errorf("buildImage: no images found")
	}
	// Merge the images together
	var horizontal, vertical int
	if len(grids) >= 2 {
		horizontal = 2
		vertical = int(math.Ceil(float64(len(grids)) / 2.0))
	} else {
		horizontal = 1
		vertical = 1
	}
	rgba, err := gim.New(grids, horizontal, vertical).Merge()
	if err != nil {
		return "", cardsWithErrors, fmt.Errorf("buildImage: unable to merge grid: %w", err)
	}
	// Create a guid to use for the image name
	guid := ksuid.New()
	fileName = fmt.Sprintf("temp_%s.png", guid.String())
	// Save the output to PNG
	file, err := os.Create(fmt.Sprintf("%s/%s", IMAGE_BASEDIR, fileName))
	if err != nil {
		return "", cardsWithErrors, fmt.Errorf("buildImage: unable to write file: %w", err)
	}
	err = png.Encode(file, rgba)
	if err != nil {
		return "", cardsWithErrors, fmt.Errorf("buildImage: unable to encode png: %w", err)
	}
	defer file.Close()
	return fileName, cardsWithErrors, nil
}

// splitCommand takes a command string (e.g., Ally:Lockjaw) and returns the filter and query
func splitCommand(s string) (filter string, query string) {
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		filter = strings.ToLower(parts[0])
		query = strings.ToLower(parts[1])
	} else {
		query = strings.ToLower(s)
	}
	return filter, query
}

// TODO - Make these take an interface, get the type of the interface and return that type?
// removeStringIndex removes a string element at a given index from the slice
func removeStringIndex(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// removeAspectIndex removes an Aspect element at a given index from the slice
func removeAspectIndex(s []*Aspect, i int) []*Aspect {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// removeHeroIndex removes a Hero element at a given index from the slice
func removeHeroIndex(s []*Hero, i int) []*Hero {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
