package embedded

import (
	"embed"
	"io/fs"
)

const (
	GameConfig     = "game_config.yaml"
	TiledConfig    = "tiled_config.yaml"
	AsepriteConfig = "aseprite_config.yaml"
	DefaultImage   = "default_10x10.png"
)

//go:embed assets
var files embed.FS

var (
	Assets  = mustSub(files, "assets") // configs/ and images/, as a resource source
	Configs = mustSub(files, "assets/configs")
	Images  = mustSub(files, "assets/images")
)

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
