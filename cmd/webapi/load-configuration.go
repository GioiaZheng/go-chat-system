package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration 描述 Web API 的配置。
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:conf\\config.yml"` // Windows 路径分隔符
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
		Filename string `conf:"default:data\\wasatext.db"` // Windows 路径分隔符
	}
}

// loadConfiguration 从环境变量、命令行参数和配置文件加载配置。
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// 从环境变量和命令行参数加载配置
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("生成配置说明失败: %w", err)
			}
			fmt.Println(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("解析配置失败: %w", err)
	}

	// 如果配置文件存在，则从 YAML 覆盖配置
	cfgPath := filepath.FromSlash(cfg.Config.Path) // 确保路径格式兼容 Windows
	fp, err := os.Open(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("配置文件存在但读取失败: %w", err)
	} else if err == nil {
		defer fp.Close()
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("读取配置文件失败: %w", err)
		}
		if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
			return cfg, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// **确保数据库目录存在**
	dbDir := filepath.Dir(cfg.DB.Filename)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return cfg, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	return cfg, nil
}
