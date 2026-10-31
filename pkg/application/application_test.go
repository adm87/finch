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
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
)

func TestNewWithNilConfigUsesDefaults(t *testing.T) {
	app, err := application.New(nil, logger.N)
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
	}{
		{"zero size", 0, 0},
		{"zero width", 0, 600},
		{"negative height", 800, -1},
		{"larger than ebiten allows", 16385, 600},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := *application.DefaultConfig()
			cfg.Screen.Width, cfg.Screen.Height = tc.width, tc.height
			app, err := application.New(&cfg, logger.N)
			require.Nil(t, app)
			require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
		})
	}
}

func TestNewCopiesConfig(t *testing.T) {
	cfg := *application.DefaultConfig()
	cfg.Screen.Width, cfg.Screen.Height = 320, 180

	app, err := application.New(&cfg, logger.N)
	require.NoError(t, err)

	cfg.Screen.Width = 1920
	got := app.Config()
	require.Equal(t, 320, got.Screen.Width)

	got.Screen.Height = 1
	require.Equal(t, 180, app.Config().Screen.Height)
}

func TestRegisterFileSystems(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	fileSystems := testFileSystems(t)
	require.NoError(t, app.RegisterFileSystems(fileSystems))

	fsys, err := app.Assets().FileSystem("test")
	require.NoError(t, err)
	require.NotNil(t, fsys)

	require.True(t, errs.IsAlreadyExists(app.RegisterFileSystems(fileSystems)))
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

	app, err := application.New(nil, logger.N)
	require.NoError(t, err)
	require.NoError(t, app.RegisterFileSystems(fileSystems))
	require.NoError(t, app.Assets().Load(fileSystems["test"], "sprite.png", "config.json"))

	imgHandle, ok := app.Cache().Images().Lookup("sprite.png")
	require.True(t, ok)
	img, err := app.Cache().Images().Get(imgHandle)
	require.NoError(t, err)
	require.Equal(t, image.Pt(3, 2), img.Bounds().Size())

	dataHandle, ok := app.Cache().Data().Lookup("config.json")
	require.True(t, ok)
	raw, err := app.Cache().Data().Get(dataHandle)
	require.NoError(t, err)
	require.JSONEq(t, `{"ok": true}`, string(raw))
}

func TestNewDefaultsNilLogger(t *testing.T) {
	app, err := application.New(nil, nil)
	require.NoError(t, err)
	require.Equal(t, logger.N, app.Logger())
}
