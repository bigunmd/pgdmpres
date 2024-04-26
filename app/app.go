package app

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

func Run() {
	// ________________________________________________________________________
	// Parse cli args to config
	filepath := pflag.StringP("config", "c", "", "configuration filepath (default: None)")
	pflag.Parse()
	if err := initCfg(*filepath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	lvl, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse logger level")
	}
	log.Logger = log.Level(lvl)
	log.Info().Msgf("Application logger level is set to '%s'", log.Logger.GetLevel().String())

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// This blocks the main thread until an interrupt is received
	<-quit

	log.Info().Msg("Gracefully shutting down")
	log.Info().Msg("Gracefull shutdown completed")
}
