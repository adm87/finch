package assets

import "io/fs"

type FileSystem string

func (f FileSystem) String() string {
	return string(f)
}

type AssetSystem struct {
	fileSystems map[FileSystem]fs.FS
}

func NewAssetSystem() *AssetSystem {
	return &AssetSystem{
		fileSystems: make(map[FileSystem]fs.FS),
	}
}

func (as *AssetSystem) RegisterFileSystem(name FileSystem, fsys fs.FS) error {
	as.fileSystems[name] = fsys
	return nil
}
