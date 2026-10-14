package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/types"
)

type GameConfig struct {
	Window    WindowConfig             `yaml:"window"`
	Rendering RenderingConfig          `yaml:"rendering"`
	Resources resources.ResourceConfig `yaml:"resources"`
}

type WindowConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
}

type RenderingConfig struct {
	ClearColor types.HexColor `yaml:"clear_color"`
}

func MustLoadConfig(root string, cfg *GameConfig) {
	filePath := filepath.Join(root, "game_config.yaml")

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("failed to unmarshal config file: %v", err)
	}
}
