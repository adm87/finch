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
