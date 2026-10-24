package application_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
)

func TestNewWithNilConfigUsesDefaults(t *testing.T) {
	app, err := application.New(nil)
	require.NoError(t, err)
	require.Equal(t, *application.DefaultConfig(), app.Config())

	w, h := app.Screen().Buffer().Bounds().Dx(), app.Screen().Buffer().Bounds().Dy()
	require.Equal(t, 800, w)
	require.Equal(t, 600, h)
}

func TestNewRejectsInvalidScreenSize(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
	}{
		{"zero", 0, 0},
		{"zero width", 0, 600},
		{"negative height", 800, -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := application.Config{Screen: application.ScreenConfig{Width: tc.width, Height: tc.height}}
			app, err := application.New(&cfg)
			require.Nil(t, app)
			require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
		})
	}
}

func TestNewCopiesConfig(t *testing.T) {
	cfg := application.Config{Screen: application.ScreenConfig{Width: 320, Height: 180}}
	app, err := application.New(&cfg)
	require.NoError(t, err)

	cfg.Screen.Width = 1920

	got := app.Config()
	require.Equal(t, 320, got.Screen.Width)

	got.Screen.Height = 1
	require.Equal(t, 180, app.Config().Screen.Height)
}

func TestNewRegistersImporters(t *testing.T) {
	var pngData bytes.Buffer
	require.NoError(t, png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 3, 2))))

	fsys := fstest.MapFS{
		"sprite.png":  {Data: pngData.Bytes()},
		"config.json": {Data: []byte(`{"ok": true}`)},
	}

	app, err := application.New(nil)
	require.NoError(t, err)
	require.NoError(t, app.Assets().Load(fsys, "sprite.png", "config.json"))

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
