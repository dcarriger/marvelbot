package server

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"marvelbot/pkg/rule"
)

// ReadRules parses rules in YAML format and returns them to the caller.
func ReadRules(path string) (rules []*rule.Rule, err error) {
	files, err := walkConfigs(path, ".yaml")
	if err != nil {
		return nil, err
	}

	// Iterate over the files and unmarshal the underlying YAML data
	for _, f := range files {
		unmarshaledRule := &rule.Rule{}
		yamlFile, err := ioutil.ReadFile(fmt.Sprintf("%s", f))
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s: %w", f, err)
		}
		err = yaml.Unmarshal(yamlFile, &unmarshaledRule)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", f, err)
		}
		rules = append(rules, unmarshaledRule)
	}
	return
}
