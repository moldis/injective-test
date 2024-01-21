// Package config provides functions for handling configuration files.
package config

import (
	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
)

// Config represents the configuration settings for the application. It contains the following fields:
// - LogLevel: The log level for the application. It can be set using the "LOG_LEVEL" environment variable. The default value is "info".
// - Listen: The address and port to listen on. It can be set using the "LISTEN" environment variable. The default value is "0.0.0.0:8080".
type Config struct {
	LogLevel      string `env:"LOG_LEVEL" envDefault:"debug"`
	Listen        string `env:"LISTEN" envDefault:"0.0.0.0:8080"`
	FetchInterval int    `env:"FETCH_INTERVAL" envDefault:"5"`
}

// New initializes a new instance of Config and performs some setup tasks.
// It sets up the error stack marshaler, creates a new logger with specified options,
// sets the time field format, and parses environment variables to populate the Config struct.
// It also sets the log level globally based on the LogLevel field in Config.
// Finally, it returns a pointer to the populated Config and any error encountered during the process.
func New() (*Config, error) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Stack().Logger()
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	zerolog.SetGlobalLevel(logLevel)

	return &cfg, nil
}
