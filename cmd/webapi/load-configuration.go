package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration describes the web API configuration.
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:./conf/config.yml"`
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
		Filename string `conf:"default:./data/wasatext.db"`
	}
}

// loadConfiguration creates a WebAPIConfiguration starting from flags, environment variables, and configuration file.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Load configuration from environment variables and command line switches
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Override values from YAML if specified and if it exists
	fp, err := os.Open(cfg.Config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，不需要报错，只是跳过
			return cfg, nil
		}
		return cfg, fmt.Errorf("can't read the config file: %w", err)
	}
	defer func() {
		_ = fp.Close()
	}()

	// 读取 YAML 配置
	yamlFile, err := ioutil.ReadAll(fp)
	if err != nil {
		return cfg, fmt.Errorf("can't read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
	}

	return cfg, nil
}
