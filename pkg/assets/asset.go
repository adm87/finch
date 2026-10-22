package assets

import (
	"path/filepath"
	"strings"
)

type Asset string

func (a Asset) String() string {
	return string(a)
}

func (a Asset) Type() string {
	return normalizeType(filepath.Ext(string(a)))
}

func normalizeType(resourceType string) string {
	return strings.ToLower(strings.Trim(resourceType, "."))
}

type AssetBundle struct {
	FileSystem FileSystem `yaml:"filesystem"`
	Assets     []Asset    `yaml:"assets"`
}

func (ab AssetBundle) Load(as *Assets) error {
	filesystem, err := as.GetFileSystem(ab.FileSystem)
	if err != nil {
		return err
	}
	return as.Load(filesystem, ab.Assets...)
}

func (ab AssetBundle) Unload(as *Assets) error {
	return as.Unload(ab.Assets...)
}
