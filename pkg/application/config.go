package application

import (
	"errors"
	"slices"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/hexcolor"
)

// maxScreenSize is the largest image ebiten can create.
const maxScreenSize = 16384

type Config struct {
	Title    string               `yaml:"title"`
	Bundles  []assets.AssetBundle `yaml:"bundles"`
	Screen   ScreenConfig         `yaml:"screen"`
	TickRate int                  `yaml:"tick_rate"`
}

type ScreenConfig struct {
	Width      int            `yaml:"width"`
	Height     int            `yaml:"height"`
	Fullscreen bool           `yaml:"fullscreen"`
	ClearColor hexcolor.Color `yaml:"clear_color"`
}

func DefaultConfig() *Config {
	return &Config{
		TickRate: 60,
		Screen: ScreenConfig{
			Width:      800,
			Height:     600,
			ClearColor: hexcolor.Color{A: 255},
		},
	}
}

func (c Config) Validate() error {
	var errList []error
	if c.TickRate <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("tick_rate", c.TickRate, "must be positive"))
	}
	errList = append(errList,
		errs.CheckRange("screen.width", c.Screen.Width, 1, maxScreenSize),
		errs.CheckRange("screen.height", c.Screen.Height, 1, maxScreenSize),
	)
	return errors.Join(errList...)
}

func (c Config) clone() Config {
	c.Bundles = slices.Clone(c.Bundles)
	for i := range c.Bundles {
		c.Bundles[i].Assets = slices.Clone(c.Bundles[i].Assets)
	}
	return c
}
