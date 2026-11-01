package display

import (
	"errors"

	"github.com/adm87/finch/pkg/errs"
)

// maxScreenSize is the largest image ebiten can create.
const maxScreenSize = 16384

// Config is the size of the buffer the game draws into, not the window.
type Config struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

func DefaultConfig() *Config {
	return &Config{Width: 800, Height: 600}
}

func (c Config) Validate() error {
	return errors.Join(
		errs.CheckRange("screen.width", c.Width, 1, maxScreenSize),
		errs.CheckRange("screen.height", c.Height, 1, maxScreenSize),
	)
}
