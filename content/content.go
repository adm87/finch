package content

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

type FileSystems map[assets.FileSystem]fs.FS

const (
	EmbeddedFS assets.FileSystem = "embedded"
	AssetsFS   assets.FileSystem = "assets"
)

const (
	AsepriteConfig    assets.Asset = "embedded/configs/aseprite_config.yaml"
	TiledConfig       assets.Asset = "embedded/configs/tiled_config.yaml"
	GameConfig        assets.Asset = "embedded/configs/game_config.yaml"
	DefaultImage10x10 assets.Asset = "embedded/images/default_10x10.png"
)

//go:embed embedded
var embeddedFiles embed.FS

var FileSystemRegistry = FileSystems{
	EmbeddedFS: embeddedFiles,
}

func InitializeFileSystems() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	contentDir := filepath.Join(root, "content")

	FileSystemRegistry[AssetsFS] = os.DirFS(filepath.Join(contentDir, string(AssetsFS)))
	return nil
}

func GetFileSystem(id assets.FileSystem) (fs.FS, error) {
	fs, ok := FileSystemRegistry[id]
	if !ok {
		return nil, errs.NotFound("file system", id.String())
	}
	return fs, nil
}
