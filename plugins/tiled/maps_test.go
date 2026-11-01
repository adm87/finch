package tiled_test

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/plugins/tiled"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// blankPNG is a 32x16 image, so tests don't need real art.
var blankPNG = func() []byte {
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 32, 16))); err != nil {
		panic(err)
	}
	return raw.Bytes()
}()

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
	"tilesets/terrain.png": {Data: blankPNG},
	"tilesets/crates.png":  {Data: blankPNG},
	"maps/props/rock.png":  {Data: blankPNG},
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

var mapImages = []string{"tilesets/terrain.png", "tilesets/crates.png", "maps/props/rock.png"}

func newMaps(t *testing.T) *tiled.Plugin {
	t.Helper()
	cfg := &tiled.Config{Maps: map[string]tiled.MapConfig{
		"level":    {TMX: "maps/level.tmx"},
		"broken":   {TMX: "broken/level.tmx"},
		"layered":  {TMX: "layered/map.tmx"},
		"infinite": {TMX: "infinite/map.tmx"},
	}}
	p, err := tiled.NewPlugin(cfg, mapFS, rendering.NewPipeline(), logger.NewLogger(io.Discard, logger.Options{}))
	require.NoError(t, err)
	return p
}

func TestLoadLoadsEverythingItRefersTo(t *testing.T) {
	maps := newMaps(t)
	ic := maps.Images()

	handle, err := maps.Load("level")
	require.NoError(t, err)
	for _, f := range mapImages {
		_, ok := ic.Lookup(f)
		require.True(t, ok, f)
	}

	again, err := maps.Load("level")
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

func TestUnloadForgetsTheMap(t *testing.T) {
	maps := newMaps(t)

	handle, err := maps.Load("level")
	require.NoError(t, err)
	require.NoError(t, maps.Unload("level"))

	_, err = maps.Map(handle)
	require.True(t, errs.IsNotFound(err))
	require.True(t, errs.IsNotFound(maps.Unload("level")))
}

func TestLoadFailureKeepsNothing(t *testing.T) {
	maps := newMaps(t)

	_, err := maps.Load("broken")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.True(t, errs.IsNotFound(maps.Unload("broken")))
}
