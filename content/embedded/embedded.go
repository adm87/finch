package embedded

import (
	"embed"
	"io/fs"
)

const (
	DefaultImage = "default_10x10.png"
)

//go:embed assets
var files embed.FS

var (
	Images  = mustSub(files, "assets/images")
	Configs = mustSub(files, "assets/configs")
)

func mustSub(fsys fs.FS, dir string) fs.FS {
	subFS, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return subFS
}
