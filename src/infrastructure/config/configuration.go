package config

import (
	"bytes"
	"os"
	src "tsib/quickstart"

	"github.com/rs/zerolog"
	"go.uber.org/config"
)

// Configuration defines the overall struct for what is expected in yaml
type Configuration struct {
	Log struct {
		DefaultLevel  int
		PackagesLevel map[string]int
	}

	Server struct {
		Port    string
		Origins []string
	}
}

// constructor that takes "application.yaml" file and loads it into
// Configuration struct.
func NewConfiguration() (*Configuration, error) {
	var cfg Configuration

	yaml, err := config.NewYAML(config.Source(bytes.NewReader(src.ApplicationYaml)), config.Expand(os.LookupEnv))
	if err != nil {
		return nil, err
	}

	if err := yaml.Get(config.Root).Populate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Configuration) Print(logger zerolog.Logger) {
	logger.Info().Msgf("%+v", *cfg)
}
