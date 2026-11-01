package assets

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
)

// LoadMode controls whether a Lease may load missing assets.
type LoadMode uint8

const (
	// LoadModeRequired fails if any asset isn't already loaded.
	LoadModeRequired LoadMode = iota
	// LoadModeDynamic loads the assets that aren't loaded.
	LoadModeDynamic
)

// UnloadMode controls whether releasing a Lease unloads what it loaded.
type UnloadMode uint8

const (
	// UnloadModeKeep leaves the assets loaded.
	UnloadModeKeep UnloadMode = iota
	// UnloadModeIfOwned unloads the assets the lease loaded once nothing else
	// uses them.
	UnloadModeIfOwned
)

// Lease acquires assets for one owner so they can be released together.
type Lease struct {
	assets *Assets
	fsys   fs.FS
	held   []Asset
	mode   LoadMode
}

// Lease starts a lease that loads missing assets from fsys, if mode allows.
func (a *Assets) Lease(fsys fs.FS, mode LoadMode) *Lease {
	return &Lease{assets: a, fsys: fsys, mode: mode}
}

// Require acquires the assets in batch, or fails in LoadModeRequired if any
// isn't loaded.
func (l *Lease) Require(batch ...Asset) error {
	if l.mode == LoadModeRequired {
		var missing []Asset
		for _, asset := range batch {
			if !l.assets.Loaded(asset) {
				missing = append(missing, asset)
			}
		}
		if len(missing) > 0 {
			return errs.InvalidOperation(fmt.Sprintf("%v not loaded and mode is LoadModeRequired", missing))
		}
	}

	if err := l.assets.Acquire(l.fsys, batch...); err != nil {
		return err
	}
	l.held = append(l.held, batch...)
	return nil
}

// Release releases everything the lease acquired. Releasing again does nothing.
func (l *Lease) Release(mode UnloadMode) error {
	if mode != UnloadModeIfOwned {
		l.assets.Keep(l.held...)
	}
	held := l.held
	l.held = nil
	return l.assets.Release(held...)
}
