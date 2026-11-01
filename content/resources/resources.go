package resources

import (
	"io/fs"
	"os"
	"path/filepath"
)

// New reads from root/assets, with a folder per tool. Nothing is opened until
// a file is loaded.
func New(root string) fs.FS {
	return os.DirFS(filepath.Join(root, "assets"))
}
