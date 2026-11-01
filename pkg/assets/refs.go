package assets

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
)

type usage struct {
	refs     int
	acquired bool // loaded by Acquire, so Release may unload it
}

// Acquire loads whichever assets in batch aren't loaded, then counts a reference
// to each one. If anything fails, it unloads what it loaded and counts nothing.
func (a *Assets) Acquire(fsys fs.FS, batch ...Asset) error {
	var missing []Asset
	for _, asset := range a.distinct(batch) {
		if !a.Loaded(asset) {
			missing = append(missing, asset)
		}
	}

	if len(missing) > 0 {
		if err := a.Load(fsys, missing...); err != nil {
			var loaded []Asset
			for _, asset := range missing {
				if a.Loaded(asset) {
					loaded = append(loaded, asset)
				}
			}
			return errors.Join(err, a.Unload(loaded...))
		}
	}

	for _, asset := range missing {
		a.usage[asset] = usage{acquired: true}
	}
	for _, asset := range batch {
		u := a.usage[asset]
		u.refs++
		a.usage[asset] = u
	}
	return nil
}

// Release drops a reference to each asset in batch. An asset Acquire loaded is
// unloaded once nothing references it.
func (a *Assets) Release(batch ...Asset) error {
	counts := make(map[Asset]int, len(batch))
	for _, asset := range batch {
		counts[asset]++
	}
	for asset, n := range counts {
		if a.usage[asset].refs < n {
			return errs.InvalidOperation(fmt.Sprintf("release %q: more releases than references", asset))
		}
	}

	var unused []Asset
	for asset, n := range counts {
		u := a.usage[asset]
		u.refs -= n
		if u.refs > 0 {
			a.usage[asset] = u
			continue
		}
		delete(a.usage, asset)
		if u.acquired {
			unused = append(unused, asset)
		}
	}
	return a.Unload(unused...)
}

// Keep stops Release from unloading the assets in batch.
func (a *Assets) Keep(batch ...Asset) {
	for _, asset := range batch {
		if u, ok := a.usage[asset]; ok {
			u.acquired = false
			a.usage[asset] = u
		}
	}
}

// References is how many times the asset has been acquired and not released.
func (a *Assets) References(asset Asset) int {
	return a.usage[asset].refs
}
