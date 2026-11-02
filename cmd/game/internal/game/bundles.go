package game

import (
	"fmt"

	"github.com/adm87/finch/pkg/assets"
)

// LoadBundles loads the bundles that stay loaded for the whole session.
func (g *Game) LoadBundles(bundles []assets.AssetBundle) error {
	for _, bundle := range bundles {
		if err := bundle.Load(g.assets); err != nil {
			return fmt.Errorf("load bundle from %q: %w", bundle.FileSystem, err)
		}
	}
	return nil
}
