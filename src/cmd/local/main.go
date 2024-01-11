package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"tsib/quickstart/infrastructure/config"
	"tsib/quickstart/infrastructure/security"
	"tsib/quickstart/infrastructure/transport"

	"github.com/ipfans/fxlogger"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func setup() {
	if err := godotenv.Load(); nil != err {
		log.Logger.Error().Err(err).Msg("unable to load environment variables")
	}
}

func main() {
	setup()
	fx.New(
		fx.Provide(
			NewRootLogger,
			config.NewConfiguration,
			transport.NewHTTPServer,
			transport.NewMuxRouter,
			transport.NewRestController,
			security.NewTokenVerifier,
		),
		fx.Supply(
			&security.TokenVerifierTransport{
				Transport: http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		),
		fx.WithLogger(func(rootLogger zerolog.Logger) fxevent.Logger {
			return fxlogger.WithZerolog(rootLogger.Level(zerolog.WarnLevel))()
		}),
		fx.Invoke(
			func(cfg *config.Configuration, rootLogger zerolog.Logger) {
				rootLogger.Info().Msg("application.yaml read")
				cfg.Print(rootLogger)
			},
			func(*http.Server) {
				// start server
			},
		),
	).Run()
}

func NewRootLogger(cfg *config.Configuration) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC1123}).
		Level(zerolog.Level(cfg.Log.DefaultLevel)).
		With().
		Timestamp().
		Caller().
		Logger()
}
