// load-configuration.go defines the configuration structures and helpers used
// to assemble runtime settings for the web API server from CLI flags,
// environment variables, and optional YAML overlays.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration captures the API server, debug, and database settings
// parsed from flags, environment variables, or a YAML configuration file.
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:/conf/config.yml"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Debug bool
	DB    struct {
		Filename string `conf:"default:/tmp/decaf.db"`
		// Filename string `conf:"default:data/wasatext.db"`
	}
}

// loadConfiguration builds a WebAPIConfiguration by parsing environment
// variables and CLI flags, then optionally layering the YAML file specified at
// Config.Path. CLI values override the environment, and the YAML file overrides
// both. The configuration file path itself is provided through the environment
// or CLI parameters.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Parse configuration from environment variables and command line switches.
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage) //nolint:forbidigo
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Apply YAML overrides if the file exists (useful in k8s/compose scenarios).
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't read the config file, while it exists: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
		_ = fp.Close()
	}

	return cfg, nil
}
