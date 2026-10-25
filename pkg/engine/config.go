package engine

import (
	"errors"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/hexcolor"
)

// Config sets up the window and the game loop.
type Config struct {
	Title      string         `yaml:"title"`
	ClearColor hexcolor.Color `yaml:"clear_color"`
	Window     WindowConfig   `yaml:"window"`
	TickRate   int            `yaml:"tick_rate"`
}

type WindowConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
	Resizable  bool `yaml:"resizable"`
}

func DefaultConfig() *Config {
	return &Config{
		ClearColor: hexcolor.Color{A: 255},
		Window:     WindowConfig{Width: 800, Height: 600},
		TickRate:   60,
	}
}

func (c Config) Validate() error {
	var errList []error
	if c.TickRate <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("tick_rate", c.TickRate, "must be positive"))
	}
	if c.Window.Width <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("window.width", c.Window.Width, "must be positive"))
	}
	if c.Window.Height <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("window.height", c.Window.Height, "must be positive"))
	}
	return errors.Join(errList...)
}
