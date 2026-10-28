package tmx

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/errs"
)

const (
	errMustNotBeEmpty = "must not be empty"
	errMustBePositive = "must be positive"
	errMustNotBeNeg   = "must not be negative"
)

// prefixed tags err with where it came from, or returns nil.
func prefixed(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", prefix, err)
}

type Orientation string

const (
	Orthogonal Orientation = "orthogonal"
	Isometric  Orientation = "isometric"
	Oblique    Orientation = "oblique"
	Staggered  Orientation = "staggered"
	Hexagonal  Orientation = "hexagonal"
)

func (o Orientation) IsValid() bool {
	switch o {
	case Orthogonal, Isometric, Oblique, Staggered, Hexagonal:
		return true
	default:
		return false
	}
}

type RenderOrder string

const (
	RightDown RenderOrder = "right-down"
	RightUp   RenderOrder = "right-up"
	LeftDown  RenderOrder = "left-down"
	LeftUp    RenderOrder = "left-up"
)

func (o RenderOrder) IsValid() bool {
	switch o {
	case RightDown, RightUp, LeftDown, LeftUp:
		return true
	default:
		return false
	}
}

type StaggerAxis string

const (
	StaggerX StaggerAxis = "x"
	StaggerY StaggerAxis = "y"
)

func (a StaggerAxis) IsValid() bool {
	return a == StaggerX || a == StaggerY
}

type StaggerIndex string

const (
	StaggerOdd  StaggerIndex = "odd"
	StaggerEven StaggerIndex = "even"
)

func (i StaggerIndex) IsValid() bool {
	return i == StaggerOdd || i == StaggerEven
}

// Map is a Tiled map file.
type Map struct {
	XMLName         xml.Name     `xml:"map"`
	Class           string       `xml:"class,attr"`
	Orientation     Orientation  `xml:"orientation,attr"`
	RenderOrder     RenderOrder  `xml:"renderorder,attr"`
	StaggerAxis     StaggerAxis  `xml:"staggeraxis,attr"`
	StaggerIndex    StaggerIndex `xml:"staggerindex,attr"`
	BackgroundColor *Color       `xml:"backgroundcolor,attr"`
	Properties      Properties   `xml:"properties>property"`
	Tilesets        []MapTileset `xml:"tileset"`
	Layers          Layers       `xml:",any"`
	Width           int          `xml:"width,attr"`
	Height          int          `xml:"height,attr"`
	TileWidth       int          `xml:"tilewidth,attr"`
	TileHeight      int          `xml:"tileheight,attr"`
	SkewX           int          `xml:"skewx,attr"`
	SkewY           int          `xml:"skewy,attr"`
	HexSideLength   int          `xml:"hexsidelength,attr"`
	ParallaxOriginX float64      `xml:"parallaxoriginx,attr"`
	ParallaxOriginY float64      `xml:"parallaxoriginy,attr"`
	Infinite        bool         `xml:"infinite,attr"`
}

func (m *Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Map
	*m = Map{RenderOrder: RightDown}
	return d.DecodeElement((*plain)(m), &start)
}

func (m *Map) Validate() error {
	var errList []error
	if !m.Orientation.IsValid() {
		errList = append(errList, errs.Argument("orientation", fmt.Sprintf("unknown orientation %q", m.Orientation)))
	}
	if !m.RenderOrder.IsValid() {
		errList = append(errList, errs.Argument("renderorder", fmt.Sprintf("unknown render order %q", m.RenderOrder)))
	}
	if m.Orientation == Staggered || m.Orientation == Hexagonal {
		if !m.StaggerAxis.IsValid() {
			errList = append(errList, errs.Argument("staggeraxis", fmt.Sprintf("unknown stagger axis %q", m.StaggerAxis)))
		}
		if !m.StaggerIndex.IsValid() {
			errList = append(errList, errs.Argument("staggerindex", fmt.Sprintf("unknown stagger index %q", m.StaggerIndex)))
		}
	}
	if m.Width <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("width", m.Width, errMustBePositive))
	}
	if m.Height <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("height", m.Height, errMustBePositive))
	}
	if m.TileWidth <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("tilewidth", m.TileWidth, errMustBePositive))
	}
	if m.TileHeight <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("tileheight", m.TileHeight, errMustBePositive))
	}
	if m.HexSideLength < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("hexsidelength", m.HexSideLength, errMustNotBeNeg))
	}
	errList = append(errList, prefixed("properties", m.Properties.Validate()))

	lastGID := 0
	for i, ts := range m.Tilesets {
		prefix := fmt.Sprintf("tileset[%d]", i)
		if ts.FirstGID <= lastGID {
			errList = append(errList, prefixed(prefix, errs.ArgumentOutOfRange("firstgid", ts.FirstGID, "must be ascending")))
		}
		lastGID = ts.FirstGID
		errList = append(errList, prefixed(prefix, ts.Validate()))
	}

	errList = append(errList, m.Layers.Validate())
	for l := range m.Layers.All() {
		if l.Tiles != nil {
			errList = append(errList, prefixed(fmt.Sprintf("layer %q", l.Tiles.Name), l.Tiles.Data.validateLayout(m.Infinite)))
		}
	}
	return errors.Join(errList...)
}
