package content

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adm87/finch/pkg/errs"
)

type FileSystemID string

func (f FileSystemID) String() string {
	return string(f)
}

type FileSystems map[FileSystemID]fs.FS

const (
	EmbeddedFS FileSystemID = "embedded"
	AssetsFS   FileSystemID = "assets"
)

const (
	AsepriteConfig    = "embedded/configs/aseprite_config.yaml"
	TiledConfig       = "embedded/configs/tiled_config.yaml"
	GameConfig        = "embedded/configs/game_config.yaml"
	DefaultImage10x10 = "embedded/images/default_10x10.png"
)

//go:embed embedded
var embeddedFiles embed.FS

var fileSystems = FileSystems{
	EmbeddedFS: embeddedFiles,
}

func InitializeFileSystems() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	contentDir := filepath.Join(root, "content")

	fileSystems[AssetsFS] = os.DirFS(filepath.Join(contentDir, string(AssetsFS)))
	return nil
}

func GetFileSystem(id FileSystemID) (fs.FS, error) {
	fs, ok := fileSystems[id]
	if !ok {
		return nil, errs.NotFound("file system", id.String())
	}
	return fs, nil
}
