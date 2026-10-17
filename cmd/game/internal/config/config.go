package config

import (
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/types"
)

type StaticContent []resources.ResourceBundle

func (sc StaticContent) Load(res resources.Resources) error {
	for _, r := range sc {
		if err := r.Load(res); err != nil {
			return err
		}
	}
	return nil
}

type GameConfig struct {
	StaticContent StaticContent   `yaml:"static_content"`
	Window        WindowConfig    `yaml:"window"`
	Rendering     RenderingConfig `yaml:"rendering"`
}

type WindowConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
}

type RenderingConfig struct {
	ClearColor types.HexColor `yaml:"clear_color"`
}
