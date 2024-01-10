package docker

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type Runtime struct {
}

func NewRuntime() *Runtime {

	return &Runtime{}
}

func (r *Runtime) PutSecretsIntoEnv() {
	folder, ok := os.LookupEnv("DOCKER_SECRETS_PATH")
	if !ok || strings.TrimSpace(folder) == "" {
		folder = "/run/secrets"
	}
	log.Info().Str("path", folder).Msg("processing docker secrets")

	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Send()
			return err
		}
		if !f.IsDir() {
			log.Debug().Str("secret", path).Msg("reading file")
			secret, err := os.ReadFile(path)
			if err != nil {
				log.Error().Err(err).Msg("ignoring secret")
			} else {
				varName := strings.ToUpper(f.Name())
				os.Setenv(varName, string(secret))
				log.Info().Str("env", varName).Msg("variable set")
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Send()
	}
}
