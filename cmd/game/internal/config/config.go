package config

import (
	"github.com/adm87/finch/pkg/types"
	"gopkg.in/yaml.v3"
)

type GameConfig struct {
	Window    WindowConfig    `yaml:"window"`
	Rendering RenderingConfig `yaml:"rendering"`
}

type WindowConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
}

type RenderingConfig struct {
	ClearColor types.HexColor `yaml:"clear_color"`
}

func LoadConfig(raw []byte) (*GameConfig, error) {
	var cfg GameConfig
	err := yaml.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
