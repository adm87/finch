package tmx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/errs"
)

type ObjectAlignment string

// AlignUnspecified means bottom-left on orthogonal maps and bottom on isometric.
const (
	AlignUnspecified ObjectAlignment = "unspecified"
	AlignTopLeft     ObjectAlignment = "topleft"
	AlignTop         ObjectAlignment = "top"
	AlignTopRight    ObjectAlignment = "topright"
	AlignLeft        ObjectAlignment = "left"
	AlignCenter      ObjectAlignment = "center"
	AlignRight       ObjectAlignment = "right"
	AlignBottomLeft  ObjectAlignment = "bottomleft"
	AlignBottom      ObjectAlignment = "bottom"
	AlignBottomRight ObjectAlignment = "bottomright"
)

func (a ObjectAlignment) IsValid() bool {
	switch a {
	case AlignUnspecified, AlignTopLeft, AlignTop, AlignTopRight, AlignLeft, AlignCenter,
		AlignRight, AlignBottomLeft, AlignBottom, AlignBottomRight:
		return true
	default:
		return false
	}
}

type TileRenderSize string

const (
	RenderSizeTile TileRenderSize = "tile"
	RenderSizeGrid TileRenderSize = "grid"
)

func (s TileRenderSize) IsValid() bool {
	return s == RenderSizeTile || s == RenderSizeGrid
}

type FillMode string

const (
	FillStretch           FillMode = "stretch"
	FillPreserveAspectFit FillMode = "preserve-aspect-fit"
)

func (m FillMode) IsValid() bool {
	return m == FillStretch || m == FillPreserveAspectFit
}

// MapTileset is a tileset entry in a map or template. It either points to a
// .tsx file or embeds the tileset.
type MapTileset struct {
	Embedded *Tileset
	Source   string // relative to the map
	FirstGID int
}

func (mt *MapTileset) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*mt = MapTileset{}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "firstgid":
			gid, err := strconv.Atoi(attr.Value)
			if err != nil {
				return errs.Format(attr.Value, "invalid firstgid").Wrap(err)
			}
			mt.FirstGID = gid
		case "source":
			mt.Source = attr.Value
		}
	}

	if mt.Source != "" {
		return d.Skip()
	}
	mt.Embedded = &Tileset{}
	return d.DecodeElement(mt.Embedded, &start)
}

func (mt *MapTileset) Validate() error {
	var errList []error
	if mt.FirstGID <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("firstgid", mt.FirstGID, errMustBePositive))
	}
	if mt.Embedded != nil {
		errList = append(errList, mt.Embedded.Validate())
	}
	return errors.Join(errList...)
}

// Tileset is a .tsx file, or a tileset embedded in a map. A tileset with an
// Image is one sheet cut into tiles, and one without is a collection where
// every tile has its own image.
type Tileset struct {
	XMLName         xml.Name        `xml:"tileset"`
	Name            string          `xml:"name,attr"`
	Class           string          `xml:"class,attr"`
	ObjectAlignment ObjectAlignment `xml:"objectalignment,attr"`
	TileRenderSize  TileRenderSize  `xml:"tilerendersize,attr"`
	FillMode        FillMode        `xml:"fillmode,attr"`
	Image           *Image          `xml:"image"`
	TileOffset      *TileOffset     `xml:"tileoffset"`
	Properties      Properties      `xml:"properties>property"`
	Tiles           []Tile          `xml:"tile"`
	TileWidth       int             `xml:"tilewidth,attr"`
	TileHeight      int             `xml:"tileheight,attr"`
	Spacing         int             `xml:"spacing,attr"`
	Margin          int             `xml:"margin,attr"`
	TileCount       int             `xml:"tilecount,attr"`
	Columns         int             `xml:"columns,attr"`
}

func (ts *Tileset) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Tileset
	*ts = Tileset{
		ObjectAlignment: AlignUnspecified,
		TileRenderSize:  RenderSizeTile,
		FillMode:        FillStretch,
	}
	return d.DecodeElement((*plain)(ts), &start)
}

// Tile finds a tile's extra data by its local ID. Tiles without any aren't listed.
func (ts *Tileset) Tile(id int) (Tile, bool) {
	for _, t := range ts.Tiles {
		if t.ID == id {
			return t, true
		}
	}
	return Tile{}, false
}

func (ts *Tileset) Validate() error {
	var errList []error
	if !ts.ObjectAlignment.IsValid() {
		errList = append(errList, errs.Argument("objectalignment", fmt.Sprintf("unknown alignment %q", ts.ObjectAlignment)))
	}
	if !ts.TileRenderSize.IsValid() {
		errList = append(errList, errs.Argument("tilerendersize", fmt.Sprintf("unknown render size %q", ts.TileRenderSize)))
	}
	if !ts.FillMode.IsValid() {
		errList = append(errList, errs.Argument("fillmode", fmt.Sprintf("unknown fill mode %q", ts.FillMode)))
	}

	// Collections store their largest tile size, which can be zero when empty.
	if ts.Image != nil {
		if ts.TileWidth <= 0 {
			errList = append(errList, errs.ArgumentOutOfRange("tilewidth", ts.TileWidth, errMustBePositive))
		}
		if ts.TileHeight <= 0 {
			errList = append(errList, errs.ArgumentOutOfRange("tileheight", ts.TileHeight, errMustBePositive))
		}
		errList = append(errList, prefixed("image", ts.Image.Validate()))
	}
	for name, v := range map[string]int{"spacing": ts.Spacing, "margin": ts.Margin, "tilecount": ts.TileCount, "columns": ts.Columns} {
		if v < 0 {
			errList = append(errList, errs.ArgumentOutOfRange(name, v, errMustNotBeNeg))
		}
	}

	errList = append(errList, prefixed("properties", ts.Properties.Validate()))
	for i, t := range ts.Tiles {
		errList = append(errList, prefixed(fmt.Sprintf("tile[%d]", i), t.Validate(ts.Image == nil)))
	}
	return errors.Join(errList...)
}

// Image is an image file, or image data embedded in the map.
type Image struct {
	Trans  *Color     `xml:"trans,attr"` // drawn transparent
	Data   *ImageData `xml:"data"`
	Format string     `xml:"format,attr"` // extension of embedded data
	Source string     `xml:"source,attr"` // relative to the file holding it
	Width  int        `xml:"width,attr"`
	Height int        `xml:"height,attr"`
}

func (img Image) Validate() error {
	var errList []error
	if img.Source == "" && img.Data == nil {
		errList = append(errList, errs.Argument("source", "needs a source or embedded data"))
	}
	if img.Data != nil {
		errList = append(errList, validateEncoding(img.Data.Encoding, img.Data.Compression))
	}
	if img.Width < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("width", img.Width, errMustNotBeNeg))
	}
	if img.Height < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("height", img.Height, errMustNotBeNeg))
	}
	return errors.Join(errList...)
}

type ImageData struct {
	Data        string          `xml:",chardata"`
	Encoding    DataEncoding    `xml:"encoding,attr"`
	Compression DataCompression `xml:"compression,attr"`
}

// Bytes decodes the embedded image file.
func (d ImageData) Bytes() ([]byte, error) {
	if d.Encoding != EncodingBase64 {
		return nil, errs.NotSupported(fmt.Sprintf("%q encoding for image data", d.Encoding))
	}
	return decodeBase64(d.Data, d.Compression)
}

// TileOffset moves tiles when they're drawn. Positive Y is down.
type TileOffset struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

// Tile is extra data for one tile in a tileset. X, Y, Width and Height crop
// its image, and a zero size uses all of it.
type Tile struct {
	Type       string       `xml:"type,attr"` // class before Tiled 1.9
	Class      string       `xml:"class,attr"`
	Image      *Image       `xml:"image"`
	Collision  *ObjectGroup `xml:"objectgroup"`
	Properties Properties   `xml:"properties>property"`
	Animation  []Frame      `xml:"animation>frame"`
	ID         int          `xml:"id,attr"`
	X          int          `xml:"x,attr"`
	Y          int          `xml:"y,attr"`
	Width      int          `xml:"width,attr"`
	Height     int          `xml:"height,attr"`
}

// ClassName is the class, whichever attribute Tiled saved it in.
func (t Tile) ClassName() string {
	if t.Class != "" {
		return t.Class
	}
	return t.Type
}

// Validate needs to know if the tile is in a collection, where every tile has
// its own image.
func (t Tile) Validate(inCollection bool) error {
	var errList []error
	if t.ID < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("id", t.ID, errMustNotBeNeg))
	}
	for name, v := range map[string]int{"x": t.X, "y": t.Y, "width": t.Width, "height": t.Height} {
		if v < 0 {
			errList = append(errList, errs.ArgumentOutOfRange(name, v, errMustNotBeNeg))
		}
	}
	if inCollection && t.Image == nil {
		errList = append(errList, errs.Argument("image", "tiles in a collection need an image"))
	}
	if t.Image != nil {
		errList = append(errList, prefixed("image", t.Image.Validate()))
	}
	if t.Collision != nil {
		errList = append(errList, prefixed("objectgroup", t.Collision.Validate()))
	}
	for i, f := range t.Animation {
		if f.TileID < 0 {
			errList = append(errList, prefixed(fmt.Sprintf("frame[%d]", i), errs.ArgumentOutOfRange("tileid", f.TileID, errMustNotBeNeg)))
		}
		if f.Duration < 0 {
			errList = append(errList, prefixed(fmt.Sprintf("frame[%d]", i), errs.ArgumentOutOfRange("duration", f.Duration, errMustNotBeNeg)))
		}
	}
	errList = append(errList, prefixed("properties", t.Properties.Validate()))
	return errors.Join(errList...)
}

// Frame shows a tile for Duration milliseconds.
type Frame struct {
	TileID   int `xml:"tileid,attr"`
	Duration int `xml:"duration,attr"`
}
