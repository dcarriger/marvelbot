package config

// Config is the top-level configuration for marvelbot's config files.
type Config struct {
	Token string `json:"token" yaml:"token"`
}
