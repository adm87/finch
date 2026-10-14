package config

import (
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/types"
)

type StatusResources []resources.ResourceConfig

func (sr StatusResources) Load(res resources.Resources) error {
	for _, r := range sr {
		if err := r.Load(res); err != nil {
			return err
		}
	}
	return nil
}

type GameConfig struct {
	StaticResources StatusResources `yaml:"static_resources"`
	Window          WindowConfig    `yaml:"window"`
	Rendering       RenderingConfig `yaml:"rendering"`
}

type WindowConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
}

type RenderingConfig struct {
	ClearColor types.HexColor `yaml:"clear_color"`
}
