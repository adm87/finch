package resources

import (
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

type ResourceBundle struct {
	FileSystem FileSystem `yaml:"filesystem"`
	Resources  []Resource `yaml:"resources"`
}

func (rb ResourceBundle) Load(res Resources) error {
	filesystem, err := res.GetFileSystem(rb.FileSystem)
	if err != nil {
		return err
	}
	return res.Load(filesystem, rb.Resources...)
}

func (rb ResourceBundle) Unload(res Resources) error {
	return res.Unload(rb.Resources...)
}
