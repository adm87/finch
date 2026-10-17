package content

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adm87/finch/pkg/types/errs"
)

type FileSystemID string

func (f FileSystemID) String() string {
	return string(f)
}

type FileSystems map[FileSystemID]fs.FS

const (
	FS_EMBEDDED FileSystemID = "embedded"
	FS_ASSETS   FileSystemID = "assets"
)

const (
	ASEPRITE_CONFIG     = "embedded/configs/aseprite_config.yaml"
	GAME_CONFIG         = "embedded/configs/game_config.yaml"
	DEFAULT_IMAGE_10X10 = "embedded/images/default_10x10.png"
)

//go:embed embedded
var embeddedFiles embed.FS

var fileSystems = FileSystems{
	FS_EMBEDDED: embeddedFiles,
}

func InitializeFileSystems() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	contentDir := filepath.Join(root, "content")

	fileSystems[FS_ASSETS] = os.DirFS(filepath.Join(contentDir, string(FS_ASSETS)))
	return nil
}

func GetFileSystem(id FileSystemID) (fs.FS, error) {
	fs, ok := fileSystems[id]
	if !ok {
		return nil, errs.NotFound("file system", id.String())
	}
	return fs, nil
}
