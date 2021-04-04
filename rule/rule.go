package rule

type Rule struct {
	Name    string   `yaml:"name"`
	Text    string   `yaml:"rule_text"`
	Text2   string   `yaml:"rule_text_2"`
	Related []string `yaml:"related"`
}
