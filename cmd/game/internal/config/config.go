package config

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/hexcolor"
)

type GameConfig struct {
	Application application.Config `yaml:"application"`
	Rendering   RenderingConfig    `yaml:"rendering"`
}

type RenderingConfig struct {
	ClearColor hexcolor.Color `yaml:"clear_color"`
}

// Parse decodes raw over the application's defaults. Unknown keys are errors.
func Parse(raw []byte) (*GameConfig, error) {
	cfg := GameConfig{Application: *application.DefaultConfig()}

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)

	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
