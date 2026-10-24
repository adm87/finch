package application_test

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
)

func TestNewWithNilConfigUsesDefaults(t *testing.T) {
	app, err := application.New(nil, nil)
	require.NoError(t, err)
	require.Equal(t, *application.DefaultConfig(), app.Config())

	w, h := app.Screen().Buffer().Bounds().Dx(), app.Screen().Buffer().Bounds().Dy()
	require.Equal(t, 800, w)
	require.Equal(t, 600, h)
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		tickRate      int
	}{
		{"zero size", 0, 0, 60},
		{"zero width", 0, 600, 60},
		{"negative height", 800, -1, 60},
		{"larger than ebiten allows", 16385, 600, 60},
		{"zero tick rate", 800, 600, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := *application.DefaultConfig()
			cfg.Screen.Width, cfg.Screen.Height = tc.width, tc.height
			cfg.TickRate = tc.tickRate
			app, err := application.New(&cfg, nil)
			require.Nil(t, app)
			require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
		})
	}
}

func TestNewCopiesConfig(t *testing.T) {
	cfg := *application.DefaultConfig()
	cfg.Screen.Width, cfg.Screen.Height = 320, 180
	cfg.Bundles = []assets.AssetBundle{{FileSystem: "test", Assets: []assets.Asset{"config.json"}}}

	app, err := application.New(&cfg, testFileSystems(t))
	require.NoError(t, err)

	cfg.Screen.Width = 1920
	cfg.Bundles[0].Assets[0] = "changed.json"

	got := app.Config()
	require.Equal(t, 320, got.Screen.Width)
	require.Equal(t, assets.Asset("config.json"), got.Bundles[0].Assets[0])

	got.Screen.Height = 1
	got.Bundles[0].Assets[0] = "changed.json"
	require.Equal(t, 180, app.Config().Screen.Height)
	require.Equal(t, assets.Asset("config.json"), app.Config().Bundles[0].Assets[0])
}

func TestNewLoadsBundles(t *testing.T) {
	cfg := *application.DefaultConfig()
	cfg.Bundles = []assets.AssetBundle{{FileSystem: "test", Assets: []assets.Asset{"sprite.png", "config.json"}}}

	app, err := application.New(&cfg, testFileSystems(t))
	require.NoError(t, err)

	_, err = app.Assets().Handle("sprite.png")
	require.NoError(t, err)
	_, err = app.Assets().Handle("config.json")
	require.NoError(t, err)
}

func TestNewFailsOnMissingBundleContent(t *testing.T) {
	cfg := *application.DefaultConfig()
	cfg.Bundles = []assets.AssetBundle{{FileSystem: "test", Assets: []assets.Asset{"missing.json"}}}

	app, err := application.New(&cfg, testFileSystems(t))
	require.Nil(t, app)
	require.True(t, errs.IsNotFound(err))

	cfg.Bundles[0].FileSystem = "unregistered"
	app, err = application.New(&cfg, testFileSystems(t))
	require.Nil(t, app)
	require.True(t, errs.IsNotFound(err))
}

func testFileSystems(t *testing.T) map[assets.FileSystem]fs.FS {
	t.Helper()
	var pngData bytes.Buffer
	require.NoError(t, png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 3, 2))))

	return map[assets.FileSystem]fs.FS{
		"test": fstest.MapFS{
			"sprite.png":  {Data: pngData.Bytes()},
			"config.json": {Data: []byte(`{"ok": true}`)},
		},
	}
}

func TestNewRegistersImporters(t *testing.T) {
	fileSystems := testFileSystems(t)
	app, err := application.New(nil, fileSystems)
	require.NoError(t, err)
	require.NoError(t, app.Assets().Load(fileSystems["test"], "sprite.png", "config.json"))

	imgHandle, err := app.Assets().Handle("sprite.png")
	require.NoError(t, err)
	img, err := app.Cache().Images().Get(images.Handle(imgHandle))
	require.NoError(t, err)
	require.Equal(t, image.Pt(3, 2), img.Bounds().Size())

	dataHandle, err := app.Assets().Handle("config.json")
	require.NoError(t, err)
	raw, err := app.Cache().Data().Get(data.Handle(dataHandle))
	require.NoError(t, err)
	require.JSONEq(t, `{"ok": true}`, string(raw))
}
