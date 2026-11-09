package hexcolor

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"strings"
)

// Color is a colour that config files can write as "#RRGGBB" or "#RRGGBBAA".
type Color color.RGBA

// UnmarshalText lets YAML and JSON decode a Color from a hex string.
func (c *Color) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// Parse reads "#RRGGBB" or "#RRGGBBAA". The # is optional, and alpha defaults to FF.
func Parse(s string) (Color, error) {
	b, err := hex.DecodeString(strings.TrimPrefix(s, "#"))
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color %q: %w", s, err)
	}
	switch len(b) {
	case 3:
		return Color{R: b[0], G: b[1], B: b[2], A: 0xff}, nil
	case 4:
		return Color{R: b[0], G: b[1], B: b[2], A: b[3]}, nil
	}
	return Color{}, fmt.Errorf("invalid hex color %q: want 6 or 8 digits", s)
}
