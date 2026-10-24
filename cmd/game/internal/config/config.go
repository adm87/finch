package config

import (
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/hexcolor"
)

type StaticContent []assets.AssetBundle

func (sc StaticContent) Load(as *assets.Assets) error {
	for _, r := range sc {
		if err := r.Load(as); err != nil {
			return err
		}
	}
	return nil
}

type GameConfig struct {
	StaticContent StaticContent            `yaml:"static_content"`
	Window        application.ScreenConfig `yaml:"window"`
	FPS           int                      `yaml:"fps"`
	Rendering     RenderingConfig          `yaml:"rendering"`
}

type RenderingConfig struct {
	ClearColor hexcolor.Color `yaml:"clear_color"`
}
