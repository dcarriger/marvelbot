package card

import (
	"reflect"
	"testing"
)

// TODO - this is somewhat deprecated now that we are using strings to represent cost and not icons. Refactor.
func TestCard_CostIcon(t *testing.T) {
	var testCases = []struct {
		name   string
		input  *Card
		output string
	}{
		{name: "Test 0 cost icon",
			input: &Card{
				Cost: 0,
			},
			output: Zero,
		},
		{name: "Test 1 cost icon",
			input: &Card{
				Cost: 1,
			},
			output: One,
		},
		{name: "Test 2 cost icon",
			input: &Card{
				Cost: 2,
			},
			output: Two,
		},
		{name: "Test 3 cost icon",
			input: &Card{
				Cost: 3,
			},
			output: Three,
		},
		{name: "Test 4 cost icon",
			input: &Card{
				Cost: 4,
			},
			output: Four,
		},
		{name: "Test 5 cost icon",
			input: &Card{
				Cost: 5,
			},
			output: Five,
		},
		{name: "Test 6 cost icon",
			input: &Card{
				Cost: 6,
			},
			output: Six,
		},
		{name: "Test no icon",
			input: &Card{
				Cost: 7,
			},
			output: "",
		},
		{name: "Test resource",
			input: &Card{
				Cost:     0,
				TypeCode: "resource",
			},
			output: Resource,
		},
	}

	for _, tt := range testCases {
		result := tt.input.CostIcon()
		if result != tt.output {
			t.Errorf("%s: want %v, got %v", tt.name, tt.output, result)
		}
	}
}

func TestCard_EmbedString(t *testing.T) {
	var testCases = []struct {
		name   string
		input  *Card
		output string
	}{
		{name: "Standard card with a name and a URL",
			input: &Card{
				Name: "Goats",
				URL:  "https://example.local",
			},
			output: "[Goats](https://example.local)",
		},
		// TODO - maybe we should return something else
		{name: "Card with a name and no URL",
			input: &Card{
				Name: "Goats",
			},
			output: "[Goats]()",
		},
		// TODO - maybe we should return something else
		{name: "Card with a URL and no name",
			input: &Card{
				URL: "https://example.local",
			},
			output: "[](https://example.local)",
		},
	}

	for _, tt := range testCases {
		result := tt.input.EmbedString()
		if result != tt.output {
			t.Errorf("%s: want %v, got %v", tt.name, tt.output, result)
		}
	}
}

func TestCard_Normalize(t *testing.T) {
	var testCases = []struct {
		name   string
		input  *Card
		output string
	}{
		{name: "Basic card with no punctuation",
			input: &Card{
				Name: "Lockjaw",
			},
			output: "lockjaw",
		},
		{name: "Card with lots of spaces and punctuation",
			input: &Card{
				Name: "Get Behind Me!",
			},
			output: "getbehindme",
		},
		{name: "Hypothetical card with numbers",
			input: &Card{
				Name: "1 2 3 4 5!",
			},
			output: "12345",
		},
	}

	for _, tt := range testCases {
		result := tt.input.Normalize()
		if result != tt.output {
			t.Errorf("%s: want %v, got %v", tt.name, tt.output, result)
		}
	}
}

// TODO - sentinel errors
func TestCard_GetImagePath(t *testing.T) {
	var testCases = []struct {
		name   string
		input  *Card
		output string
		err    bool
	}{
		{name: "Card with no path",
			input:  &Card{},
			output: "",
			err:    true,
		},
		{name: "Card with non-PNG path",
			input: &Card{
				ImageSrc: `\/bundles\/cards\/test.jpg`,
			},
			output: "images/test.png",
			err:    false,
		},
		{name: "Card with PNG path",
			input: &Card{
				ImageSrc: `\/bundles\/cards\/test.png`,
			},
			output: "images/test.png",
			err:    false,
		},
	}

	for _, tt := range testCases {
		result, err := tt.input.GetImagePath()
		if result != tt.output {
			t.Errorf("%s: want %v, got %v", tt.name, tt.output, result)
		}
		// TODO - refactor these after sentinel error types exist
		if err != nil && tt.err == false {
			t.Errorf("%s: unexpected error: %w", tt.name, tt.err)
		}
		if err == nil && tt.err == true {
			t.Errorf("%s: missing error: %w", tt.name, tt.err)
		}
	}
}

func TestCards_SortSlice(t *testing.T) {
	testCases := []struct {
		name   string
		input  Cards
		output Cards
	}{
		{name: "Sort equal cost alphabetically",
			input: Cards{
				&Card{
					Name: "Lockjaw",
					Cost: 4,
				},
				&Card{
					Name: "Avenger's Mansion",
					Cost: 4,
				},
			},
			output: Cards{
				&Card{
					Name: "Avenger's Mansion",
					Cost: 4,
				},
				&Card{
					Name: "Lockjaw",
					Cost: 4,
				},
			},
		},
		{name: "Sort resources over non-resources",
			input: Cards{
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
			},
			output: Cards{
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
			},
		},
		{name: "Complex sorting example",
			input: Cards{
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Genius",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Uppercut",
					Cost: 3,
					TypeCode: "support",
				},
				&Card{
					Name: "Helicarrier",
					Cost: 3,
					TypeCode: "support",
				},
			},
			output: Cards{
				&Card{
					Name: "Genius",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
				&Card{
					Name: "Helicarrier",
					Cost: 3,
					TypeCode: "support",
				},
				&Card{
					Name: "Uppercut",
					Cost: 3,
					TypeCode: "support",
				},
			},
		},
		{name: "Already sorted by name",
			input: Cards{
				&Card{
					Name: "Avenger's Mansion",
					Cost: 4,
				},
				&Card{
					Name: "Lockjaw",
					Cost: 4,
				},
			},
			output: Cards{
				&Card{
					Name: "Avenger's Mansion",
					Cost: 4,
				},
				&Card{
					Name: "Lockjaw",
					Cost: 4,
				},
			},
		},
		{name: "Already sorted by resource",
			input: Cards{
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
			},
			output: Cards{
				&Card{
					Name: "The Power of Leadership",
					Cost: 0,
					TypeCode: "resource",
				},
				&Card{
					Name: "Make the Call",
					Cost: 0,
					TypeCode: "event",
				},
			},
		},
		{name: "Sort based on cost",
			input: Cards{
				&Card{
					Name: "Avengers Assemble",
					Cost: 4,
				},
				&Card{
					Name: "Make the Call",
					Cost: 0,
				},
			},
			output: Cards{
				&Card{
					Name: "Make the Call",
					Cost: 0,
				},
				&Card{
					Name: "Avengers Assemble",
					Cost: 4,
				},
			},
		},
	}

	for _, tt := range testCases {
		result := tt.input.SortSlice()
		if !reflect.DeepEqual(result, tt.output) {
			t.Errorf("%s: want %v, got %v", tt.name, tt.output, result)
		}
	}
}

// TODO - good way to test this function? seems like it needs refactoring
func TestCard_DownloadImages(t *testing.T) {
	var testCases = []struct {
		name   string
		input  Card
		err    bool
	}{
		{name: "Card with no image source",
			input:  Card{},
			err:    true,
		},
	}

	for _, tt := range testCases {
		err := tt.input.DownloadImages()
		// TODO - refactor these after sentinel error types exist
		if err != nil && tt.err == false {
			t.Errorf("%s: unexpected error: %w", tt.name, tt.err)
		}
		if err == nil && tt.err == true {
			t.Errorf("%s: missing error: %w", tt.name, tt.err)
		}
	}
}