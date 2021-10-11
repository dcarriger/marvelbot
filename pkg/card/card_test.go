package card

import (
	"testing"
)

// TODO - good way to test this function? seems like it needs refactoring
func TestCard_DownloadImages(t *testing.T) {
	var testCases = []struct {
		name  string
		input Card
		err   bool
	}{
		{name: "Card with no image source",
			input: Card{},
			err:   true,
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
