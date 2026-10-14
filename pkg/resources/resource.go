package resources

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type Resource string

func (r Resource) String() string {
	return string(r)
}

func (r Resource) Type() string {
	return normalizeType(filepath.Ext(string(r)))
}

func normalizeType(resourceType string) string {
	return strings.ToLower(strings.Trim(resourceType, "."))
}

type ResourceBundle []Resource

func (rb ResourceBundle) Load(res Resources, filesystem fs.FS) error {
	return res.Load(filesystem, rb...)
}

func (rb ResourceBundle) Unload(res Resources) error {
	return res.Unload(rb...)
}

type ResourceConfig struct {
	FileSystem FileSystem     `yaml:"filesystem"`
	Bundle     ResourceBundle `yaml:"bundle"`
}

func (rc ResourceConfig) Load(res Resources) error {
	filesystem, err := res.GetFileSystem(rc.FileSystem)
	if err != nil {
		return err
	}
	return rc.Bundle.Load(res, filesystem)
}

func (rc ResourceConfig) Unload(res Resources) error {
	return rc.Bundle.Unload(res)
}
