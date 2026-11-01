package aseprite_test

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
)

// sheetPNG is a real png, so the app's own image importer can load it.
var sheetPNG = func() []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 32, 16))); err != nil {
		panic(err)
	}
	return buf.Bytes()
}()

var heroAtlas = []byte(`{
	"frames": [{"frame": {"x": 0, "y": 0, "w": 16, "h": 16}, "duration": 100}],
	"meta": {"frameTags": [{"name": "Idle", "direction": "forward", "from": 0, "to": 0}]}
}`)

var libraryFS = fstest.MapFS{
	"hero.json":    {Data: heroAtlas},
	"villain.json": {Data: heroAtlas},
	"sheet.png":    {Data: sheetPNG},
}

func newPlugin(t *testing.T) (*aseprite.Plugin, *assets.Assets) {
	t.Helper()
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)
	require.NoError(t, app.RegisterFileSystems(map[assets.FileSystem]fs.FS{"test": libraryFS}))

	plugin := aseprite.NewPlugin(&aseprite.Config{Libraries: map[string]aseprite.LibraryConfig{
		"hero":    {FileSystem: "test", Atlas: "hero.json", Image: "sheet.png"},
		"villain": {FileSystem: "test", Atlas: "villain.json", Image: "sheet.png"},
	}})
	require.NoError(t, app.RegisterPlugin(plugin))
	return plugin, app.Assets()
}

func TestClearLibraryUnloadsWhatItLoaded(t *testing.T) {
	plugin, as := newPlugin(t)

	handle, err := plugin.CreateLibrary("hero", assets.LoadModeDynamic)
	require.NoError(t, err)
	require.NoError(t, plugin.ClearLibrary("hero", assets.UnloadModeIfOwned))

	require.False(t, as.Loaded("hero.json"))
	require.False(t, as.Loaded("sheet.png"))
	_, err = plugin.Library(handle)
	require.Error(t, err)
}

func TestClearLibraryKeepsFilesAnotherLibraryUses(t *testing.T) {
	plugin, as := newPlugin(t)

	_, err := plugin.CreateLibrary("hero", assets.LoadModeDynamic)
	require.NoError(t, err)
	_, err = plugin.CreateLibrary("villain", assets.LoadModeDynamic)
	require.NoError(t, err)

	require.NoError(t, plugin.ClearLibrary("hero", assets.UnloadModeIfOwned))
	require.False(t, as.Loaded("hero.json"))
	require.True(t, as.Loaded("sheet.png"))

	require.NoError(t, plugin.ClearLibrary("villain", assets.UnloadModeIfOwned))
	require.False(t, as.Loaded("sheet.png"))
}

func TestClearLibraryKeepsFilesLoadedBeforeIt(t *testing.T) {
	plugin, as := newPlugin(t)
	require.NoError(t, as.Load(libraryFS, "sheet.png"))

	_, err := plugin.CreateLibrary("hero", assets.LoadModeDynamic)
	require.NoError(t, err)
	require.NoError(t, plugin.ClearLibrary("hero", assets.UnloadModeIfOwned))

	require.True(t, as.Loaded("sheet.png"))
	require.False(t, as.Loaded("hero.json"))
}

func TestCreateLibraryRequiredNeedsFilesLoaded(t *testing.T) {
	plugin, as := newPlugin(t)

	_, err := plugin.CreateLibrary("hero", assets.LoadModeRequired)
	require.ErrorIs(t, err, errs.ErrInvalidOperation)
	require.False(t, as.Loaded("hero.json"))
}

func TestAdvanceAnimationsReportsAClearedLibraryOnce(t *testing.T) {
	plugin, _ := newPlugin(t)
	handle, err := plugin.CreateLibrary("hero", assets.LoadModeDynamic)
	require.NoError(t, err)

	w := world.NewWorld()
	entry := w.Create(images.ImageComponent, aseprite.AnimatorComponent)
	anim, _ := aseprite.GetAnimator(entry)
	anim.SetLibrary(handle)
	anim.SetAnimation("Idle")
	entries := []*donburi.Entry{entry}

	require.NoError(t, aseprite.AdvanceAnimations(0.1, plugin, entries))
	require.NoError(t, plugin.ClearLibrary("hero", assets.UnloadModeIfOwned))

	require.Error(t, aseprite.AdvanceAnimations(0.1, plugin, entries))
	require.Zero(t, anim.Library())
	require.NoError(t, aseprite.AdvanceAnimations(0.1, plugin, entries))
}
