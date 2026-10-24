package application

import (
	"errors"

	"github.com/adm87/finch/pkg/errs"
)

type Config struct {
	Screen ScreenConfig `yaml:"screen"`
}

type ScreenConfig struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
}

func DefaultConfig() *Config {
	return &Config{
		Screen: ScreenConfig{
			Width:  800,
			Height: 600,
		},
	}
}

func (c Config) Validate() error {
	var errList []error
	if c.Screen.Width <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("screen.width", c.Screen.Width, "must be positive"))
	}
	if c.Screen.Height <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("screen.height", c.Screen.Height, "must be positive"))
	}
	return errors.Join(errList...)
}
