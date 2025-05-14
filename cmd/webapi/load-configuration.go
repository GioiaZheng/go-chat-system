package main

import (
	"os"
	"time"

	"github.com/ardanlabs/conf"
)

// WebAPIConfiguration holds configuration options loaded from environment variables or flags.
type WebAPIConfiguration struct {
	Web struct {
		APIHost         string        `conf:"default:localhost:8080"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		ShutdownTimeout time.Duration `conf:"default:20s"`
	}

	DB struct {
		Filename string `conf:"default:wasa.sqlite"`
	}

	Debug bool `conf:"default:false"`
}

// loadConfiguration parses configuration from environment variables and command-line flags.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration
	if err := conf.Parse(os.Args[1:], "WASA", &cfg); err != nil {
		return WebAPIConfiguration{}, err
	}
	return cfg, nil
}
