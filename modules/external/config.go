package external

import (
	"gopkg.in/yaml.v3"
)

// Config contains the configuration about the actions module
type Config struct {
	URL string `yaml:"url"`
}

// NewConfig returns a new Config instance
func NewConfig(url string) *Config {
	return &Config{
		URL: url,
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		URL: "http://localhost:3001/api/v1/handleTx",
	}
}

func ParseConfig(bz []byte) (*Config, error) {
	type T struct {
		Config *Config `yaml:"external"`
	}
	var cfg T
	err := yaml.Unmarshal(bz, &cfg)
	return cfg.Config, err
}
