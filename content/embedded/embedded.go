package embedded

import (
	"embed"
	"io/fs"
)

const GameConfigPath = "configs/game.yaml"

//go:embed assets
var files embed.FS

var Assets = must(files, "assets")

func must(fsys fs.FS, path string) fs.FS {
	sub, err := fs.Sub(fsys, path)
	if err != nil {
		panic(err)
	}
	return sub
}
