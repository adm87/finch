package tmx

import (
	"encoding/xml"
	"errors"

	"github.com/adm87/finch/pkg/errs"
)

// Template is a .tx object template. Tile objects also reference their
// tileset, which has to be an external file.
type Template struct {
	XMLName xml.Name    `xml:"template"`
	Tileset *MapTileset `xml:"tileset"`
	Object  Object      `xml:"object"`
}

func (t Template) Validate() error {
	var errList []error
	if t.Tileset != nil {
		if t.Tileset.Embedded != nil {
			errList = append(errList, errs.NotSupported("embedded tilesets in templates"))
		}
		errList = append(errList, prefixed("tileset", t.Tileset.Validate()))
	}
	errList = append(errList, prefixed("object", t.Object.Validate()))
	return errors.Join(errList...)
}
