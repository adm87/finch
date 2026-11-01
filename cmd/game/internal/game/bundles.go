package game

import (
	"fmt"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/resource"
)

// LoadBundles loads each bundle's images, maps and libraries from its source.
// Loading a bundle twice reuses what's already loaded.
func (g *Game) LoadBundles(names ...string) error {
	for _, name := range names {
		if err := g.loadBundle(name); err != nil {
			return fmt.Errorf("bundle %q: %w", name, err)
		}
	}
	return nil
}

func (g *Game) loadBundle(name string) error {
	bundle, ok := g.bundles[name]
	if !ok {
		return errs.NotFound("bundle", name)
	}

	fsys, err := g.sources.FS(bundle.Source)
	if err != nil {
		return err
	}
	for _, img := range bundle.Images {
		if _, err := g.images.Load(resource.Key(bundle.Source, img), fsys, img); err != nil {
			return err
		}
	}
	for _, m := range bundle.Maps {
		if _, err := g.tiled.Load(bundle.Source, m); err != nil {
			return err
		}
	}
	for _, lib := range bundle.Libraries {
		if _, err := g.aseprite.Load(bundle.Source, lib); err != nil {
			return err
		}
	}
	return nil
}
