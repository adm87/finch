package resources

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Resources are the asset folders read from disk, one per tool.
type Resources struct {
	Aseprite fs.FS
	Tiled    fs.FS
}

// New reads from root/assets. Nothing is opened until a file is loaded.
func New(root string) Resources {
	assets := filepath.Join(root, "assets")
	return Resources{
		Aseprite: os.DirFS(filepath.Join(assets, "aseprite")),
		Tiled:    os.DirFS(filepath.Join(assets, "tiled")),
	}
}
