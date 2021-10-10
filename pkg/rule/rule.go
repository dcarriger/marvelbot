package rule

// Rule contains the text of a Marvel Champions rule as defined in the
// Rules Reference Guide.
type Rule struct {
	Name    string   `yaml:"name,omitempty"`
	Version string   `yaml:"version,omitempty"`
	Text    string   `yaml:"rule_text,omitempty"`
	Text2   string   `yaml:"rule_text_2,omitempty"`
	Related []string `yaml:"related,omitempty"`
}
