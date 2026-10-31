package tiled_test

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/tiled"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// blankImporter stands in for the png importer, so tests don't need real images.
type blankImporter struct {
	cache *images.Cache
}

func (b blankImporter) Import(asset assets.Asset, _ fs.FS, _ io.Reader) error {
	_, err := b.cache.Put(asset.String(), ebiten.NewImage(32, 16))
	return err
}

func (b blankImporter) Delete(asset assets.Asset) error {
	handle, exists := b.cache.Lookup(asset.String())
	if !exists {
		return errs.NotFound("image", asset.String())
	}
	_, err := b.cache.Delete(handle)
	return err
}

func (b blankImporter) Lookup(name string) bool {
	_, exists := b.cache.Lookup(name)
	return exists
}

var mapFS = fstest.MapFS{
	"maps/level.tmx": {Data: []byte(`<map orientation="orthogonal" width="2" height="1" tilewidth="16" tileheight="16">
 <tileset firstgid="1" source="../tilesets/terrain.tsx"/>
 <tileset firstgid="3" name="props" tilewidth="16" tileheight="16" tilecount="1" columns="0">
  <tile id="0"><image source="props/rock.png" width="16" height="16"/></tile>
 </tileset>
 <layer id="1" name="ground" width="2" height="1"><data encoding="csv">2,3</data></layer>
 <objectgroup id="2" name="things">
  <object id="1" template="crate.tx" x="0" y="0"/>
  <object id="2" template="crate.tx" x="16" y="0"/>
 </objectgroup>
</map>`)},
	"maps/crate.tx": {Data: []byte(`<template>
 <tileset firstgid="1" source="../tilesets/crates.tsx"/>
 <object gid="1" width="16" height="16"/>
</template>`)},
	"tilesets/terrain.tsx": {Data: []byte(`<tileset name="terrain" tilewidth="16" tileheight="16" tilecount="2" columns="2">
 <image source="terrain.png" width="32" height="16"/>
</tileset>`)},
	"tilesets/crates.tsx": {Data: []byte(`<tileset name="crates" tilewidth="16" tileheight="16" tilecount="2" columns="2">
 <image source="crates.png" width="32" height="16"/>
</tileset>`)},
	"tilesets/terrain.png": {},
	"tilesets/crates.png":  {},
	"maps/props/rock.png":  {},
	"broken/level.tmx": {Data: []byte(`<map orientation="orthogonal" width="1" height="1" tilewidth="16" tileheight="16">
 <tileset firstgid="1" source="missing.tsx"/>
 <layer id="1" name="ground" width="1" height="1"><data encoding="csv">1</data></layer>
</map>`)},
	"layered/map.tmx": {Data: []byte(`<map orientation="orthogonal" width="2" height="2" tilewidth="16" tileheight="16">
 <tileset firstgid="1" source="../tilesets/terrain.tsx"/>
 <layer id="1" name="ground" width="2" height="2"><data encoding="csv">1,1,1,1</data></layer>
 <objectgroup id="2" name="player"/>
 <group id="3" name="above" offsetx="8" offsety="4" opacity="0.5">
  <layer id="4" name="roofs" width="2" height="2" offsetx="2">
   <properties><property name="depth" type="int" value="50"/></properties>
   <data encoding="csv">2,0,0,2</data>
  </layer>
  <layer id="5" name="hidden" width="2" height="2" visible="0"><data encoding="csv">1,1,1,1</data></layer>
 </group>
 <objectgroup id="6" name="shadow" offsety="1">
  <properties><property name="depth" type="int" value="-3"/></properties>
 </objectgroup>
</map>`)},
	"infinite/map.tmx": {Data: []byte(`<map orientation="orthogonal" width="4" height="4" tilewidth="16" tileheight="16" infinite="1">
 <tileset firstgid="1" source="../tilesets/terrain.tsx"/>
 <layer id="1" name="ground" width="4" height="4">
  <data encoding="csv">
   <chunk x="-6" y="-6" width="2" height="2">1,0,0,1</chunk>
   <chunk x="-4" y="-3" width="2" height="2">2,2,0,2</chunk>
  </data>
 </layer>
</map>`)},
}

var mapFiles = []assets.Asset{
	"maps/level.tmx", "maps/crate.tx",
	"tilesets/terrain.tsx", "tilesets/terrain.png",
	"tilesets/crates.tsx", "tilesets/crates.png",
	"maps/props/rock.png",
}

func newMaps(t *testing.T) (*tiled.Maps, *assets.Assets, *images.Cache) {
	t.Helper()
	as := assets.NewAssets()
	dc, ic := data.NewCache(), images.NewCache()
	require.NoError(t, as.RegisterFileSystem("test", mapFS))
	require.NoError(t, as.RegisterImporter(data.NewImporter(dc), []string{"tmx", "tsx", "tx"}))
	require.NoError(t, as.RegisterImporter(blankImporter{cache: ic}, []string{"png"}))

	cfg := &tiled.Config{Maps: map[string]tiled.MapConfig{
		"level":    {FileSystem: "test", TMX: "maps/level.tmx"},
		"broken":   {FileSystem: "test", TMX: "broken/level.tmx"},
		"layered":  {FileSystem: "test", TMX: "layered/map.tmx"},
		"infinite": {FileSystem: "test", TMX: "infinite/map.tmx"},
	}}
	return tiled.NewMaps(as, dc, ic, cfg, logger.NewLogger(io.Discard, logger.Options{})), as, ic
}

func TestCreateMapLoadsEverythingItRefersTo(t *testing.T) {
	maps, as, ic := newMaps(t)

	handle, err := maps.CreateMap("level", tiled.LoadModeDynamic)
	require.NoError(t, err)
	for _, f := range mapFiles {
		require.True(t, as.Loaded(f), f)
	}

	again, err := maps.CreateMap("level", tiled.LoadModeDynamic)
	require.NoError(t, err)
	require.Equal(t, handle, again)

	info, err := maps.Map(handle)
	require.NoError(t, err)
	require.Len(t, info.Tilesets, 2)
	require.Equal(t, "tilesets/terrain.tsx", info.Map.Tilesets[0].Source)
	require.Len(t, info.Layers, 1)
	require.Equal(t, []tmx.GID{2, 3}, info.Layers[0].GIDs)

	// Tile 2 is the second frame of the terrain sheet.
	ts, id, ok := info.Tileset(info.Layers[0].GIDs[0])
	require.True(t, ok)
	require.Equal(t, "terrain", ts.Tileset.Name)
	img, frame, ok := ts.TileImage(id)
	require.True(t, ok)
	f, err := ic.Frame(img, frame)
	require.NoError(t, err)
	require.Equal(t, 16, f.Image.Bounds().Min.X)

	// Tile 3 is the collection's only tile.
	ts, id, ok = info.Tileset(info.Layers[0].GIDs[1])
	require.True(t, ok)
	require.Equal(t, "props", ts.Tileset.Name)
	rock, ok := ic.Lookup("maps/props/rock.png")
	require.True(t, ok)
	img, _, ok = ts.TileImage(id)
	require.True(t, ok)
	require.Equal(t, rock, img)

	_, _, ok = info.Tileset(0)
	require.False(t, ok)
}

func TestCreateMapRequiredNeedsFilesLoaded(t *testing.T) {
	maps, as, _ := newMaps(t)

	_, err := maps.CreateMap("level", tiled.LoadModeRequired)
	require.ErrorIs(t, err, errs.ErrInvalidOperation)

	require.NoError(t, as.Load(mapFS, mapFiles...))
	handle, err := maps.CreateMap("level", tiled.LoadModeRequired)
	require.NoError(t, err)

	// Nothing was owned, so clearing keeps the files.
	require.NoError(t, maps.ClearMap("level", tiled.UnloadModeIfOwned))
	require.True(t, as.Loaded("maps/level.tmx"))
	_, err = maps.Map(handle)
	require.True(t, errs.IsNotFound(err))
}

func TestClearMapUnloadsWhatItLoaded(t *testing.T) {
	maps, as, _ := newMaps(t)

	_, err := maps.CreateMap("level", tiled.LoadModeDynamic)
	require.NoError(t, err)
	require.NoError(t, maps.ClearMap("level", tiled.UnloadModeIfOwned))

	for _, f := range mapFiles {
		require.False(t, as.Loaded(f), f)
	}
	require.True(t, errs.IsNotFound(maps.ClearMap("level", tiled.UnloadModeKeep)))
}

func TestCreateMapFailureUnloadsWhatItLoaded(t *testing.T) {
	maps, as, _ := newMaps(t)

	_, err := maps.CreateMap("broken", tiled.LoadModeDynamic)
	require.True(t, errs.IsNotFound(err))
	require.False(t, as.Loaded("broken/level.tmx"))
}
