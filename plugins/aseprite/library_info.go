package aseprite

import (
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/images"
)

// LibraryInfo is a decoded library and the handles of its content.
type LibraryInfo struct {
	Atlas       *Atlas
	lease       *assets.Lease // released when the library is cleared
	ImageHandle images.Handle
	DataHandle  data.Handle
}
