package aseprite

import (
	"encoding/json"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
)

// libraryLoader builds one library's info, leasing the files it uses.
type libraryLoader struct {
	plugin *Plugin
	lease  *assets.Lease
}

func (l *libraryLoader) load(cfg *LibraryConfig) (LibraryInfo, error) {
	if err := l.lease.Require(cfg.Assets()...); err != nil {
		return LibraryInfo{}, err
	}

	info, err := l.atlas(cfg.Atlas)
	if err != nil {
		return LibraryInfo{}, err
	}

	imageHandle, exists := l.plugin.images.Lookup(cfg.Image.String())
	if !exists {
		return LibraryInfo{}, errs.NotFound("image", cfg.Image.String())
	}
	if err := l.plugin.images.SliceFrames(imageHandle, info.Atlas.Frames.FrameRects()); err != nil {
		return LibraryInfo{}, fmt.Errorf("slice image %q: %w", cfg.Image, err)
	}
	info.ImageHandle = imageHandle
	return info, nil
}

// atlas parses and validates the atlas json.
func (l *libraryLoader) atlas(asset assets.Asset) (LibraryInfo, error) {
	path := asset.String()

	dataHandle, exists := l.plugin.data.Lookup(path)
	if !exists {
		return LibraryInfo{}, errs.NotFound("data", path)
	}
	raw, err := l.plugin.data.Get(dataHandle)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", path, err)
	}

	info := LibraryInfo{DataHandle: dataHandle}
	if err := json.Unmarshal(raw, &info.Atlas); err != nil {
		return LibraryInfo{}, errs.Format(path, "invalid aseprite json").Wrap(err)
	}
	if info.Atlas == nil {
		return LibraryInfo{}, errs.Format(path, "atlas is empty")
	}
	if dropped := info.Atlas.DropMissingTags(); len(dropped) > 0 {
		l.plugin.logger.Warn("dropping missing tags", logger.String("atlas", path), logger.Strings("tags", dropped))
	}
	if err := info.Atlas.Validate(); err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", path, err)
	}

	info.Atlas.fitTagsToFrames()
	return info, nil
}
