package content

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FileSystemID string

type FileSystems map[FileSystemID]fs.FS

const (
	FS_EMBEDDED FileSystemID = "embedded"
	FS_ASSETS   FileSystemID = "assets"
)

const (
	GAME_CONFIG         = "embedded/configs/game_config.yaml"
	DEFAULT_IMAGE_10X10 = "embedded/images/default_10x10.png"
)

//go:embed embedded
var embeddedFiles embed.FS

var fileSystems = FileSystems{
	FS_EMBEDDED: embeddedFiles,
}

func InitFileSytems() error {
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
		return nil, fmt.Errorf("file system with ID %s not found", id)
	}
	return fs, nil
}
