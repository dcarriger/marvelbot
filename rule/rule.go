package rule

type Rule struct {
	Name string `yaml:"name"`
	Text string `yaml:"rule_text"`
	Related []string `yaml:"related"`
}