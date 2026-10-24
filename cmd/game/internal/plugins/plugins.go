package plugins

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/aseprite"
)

type Plugins struct {
	Aseprite *aseprite.Libraries
}

func New(app *application.Application, logger logger.Logger) (*Plugins, error) {
	ase, err := initializeAseprite(app, logger)
	if err != nil {
		return nil, fmt.Errorf("initialize aseprite: %w", err)
	}
	return &Plugins{
		Aseprite: ase,
	}, nil
}

func initializeAseprite(app *application.Application, logger logger.Logger) (*aseprite.Libraries, error) {
	cache := app.Cache()

	handle, err := app.Assets().Handle(content.AsepriteConfig)
	if err != nil {
		return nil, err
	}
	raw, err := cache.Data().Get(data.Handle(handle))
	if err != nil {
		return nil, err
	}

	var aseConfig aseprite.Config
	if err := yaml.Unmarshal(raw, &aseConfig); err != nil {
		return nil, errs.Format(content.AsepriteConfig, "invalid yaml").Wrap(err)
	}

	ase, err := aseprite.New(app.Assets(), logger, cache.Data(), cache.Images(), &aseConfig)
	if err != nil {
		return nil, err
	}

	return ase, nil
}
