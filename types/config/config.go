package config

import (
	"github.com/forbole/bdjuno/v4/modules/external"
	initcmd "github.com/forbole/juno/v5/cmd/init"
	junoconfig "github.com/forbole/juno/v5/types/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/forbole/bdjuno/v4/modules/actions"
)

// Config represents the BDJuno configuration
type Config struct {
	JunoConfig     junoconfig.Config `yaml:"-,inline"`
	ActionsConfig  actions.Config    `yaml:"actions"`
	ExternalConfig external.Config   `yaml:"external"`
}

// NewConfig returns a new Config instance
func NewConfig(junoCfg junoconfig.Config, actionsCfg actions.Config, externalCfg external.Config) Config {
	return Config{
		JunoConfig:     junoCfg,
		ActionsConfig:  actionsCfg,
		ExternalConfig: externalCfg,
	}
}

// GetBytes implements WritableConfig
func (c Config) GetBytes() ([]byte, error) {
	return yaml.Marshal(&c)
}

// Creator represents a configuration creator
func Creator(_ *cobra.Command) initcmd.WritableConfig {
	return NewConfig(junoconfig.DefaultConfig(), *actions.DefaultConfig(), external.DefaultConfig())
}
