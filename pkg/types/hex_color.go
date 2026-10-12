package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"strings"

	"gopkg.in/yaml.v3"
)

type HexColor color.RGBA

func (c *HexColor) ToColorRGBA() color.RGBA {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: c.A,
	}
}

func (c *HexColor) parseHex(s string) error {
	s = strings.TrimPrefix(s, "#")
	b, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("invalid hex color %q: %w", s, err)
	}
	if len(b) != 4 {
		return fmt.Errorf("invalid hex color %q: expected 4 bytes (RGBA), got %d", s, len(b))
	}
	*c = HexColor{R: b[0], G: b[1], B: b[2], A: b[3]}
	return nil
}

func (c *HexColor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return c.parseHex(s)
}

func (c *HexColor) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	return c.parseHex(s)
}
