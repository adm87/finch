package tiled

import (
	"path"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

func mapRefs(m *tmx.Map, dir string) []assets.Asset {
	var refs []assets.Asset
	for i := range m.Tilesets {
		ts := &m.Tilesets[i]
		if ts.Embedded != nil {
			refs = append(refs, tilesetRefs(ts.Embedded, dir)...)
			continue
		}
		refs = appendRef(refs, dir, &ts.Source)
	}

	for l := range m.Layers.All() {
		if l.Image != nil {
			refs = appendImageRef(refs, dir, l.Image.Image)
		}
		if l.Objects == nil {
			continue
		}
		for i := range l.Objects.Objects {
			refs = appendRef(refs, dir, &l.Objects.Objects[i].Template)
		}
	}
	return refs
}

func tilesetRefs(ts *tmx.Tileset, dir string) []assets.Asset {
	refs := appendImageRef(nil, dir, ts.Image)
	for i := range ts.Tiles {
		refs = appendImageRef(refs, dir, ts.Tiles[i].Image)
	}
	return refs
}

func templateRefs(t *tmx.Template, dir string) []assets.Asset {
	if t.Tileset == nil {
		return nil
	}
	return appendRef(nil, dir, &t.Tileset.Source)
}

// appendImageRef skips images embedded as data, which have no file.
func appendImageRef(refs []assets.Asset, dir string, img *tmx.Image) []assets.Asset {
	if img == nil {
		return refs
	}
	return appendRef(refs, dir, &img.Source)
}

// appendRef rewrites ref to its resolved path, so lookups after import can use
// it as is.
func appendRef(refs []assets.Asset, dir string, ref *string) []assets.Asset {
	if *ref == "" {
		return refs
	}
	*ref = path.Join(dir, *ref)
	return append(refs, assets.Asset(*ref))
}

// distinct drops repeated refs, keeping the order they were found in.
func distinct(refs []assets.Asset) []assets.Asset {
	seen := make(map[assets.Asset]struct{}, len(refs))
	out := refs[:0]
	for _, ref := range refs {
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}
