package tmx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iter"

	"github.com/adm87/finch/pkg/errs"
)

// LayerBase holds what every kind of layer has. On a group, offset, opacity,
// visibility and tint also apply to the layers inside it.
type LayerBase struct {
	Name       string     `xml:"name,attr"`
	Class      string     `xml:"class,attr"`
	TintColor  *Color     `xml:"tintcolor,attr"`
	Properties Properties `xml:"properties>property"`
	ID         int        `xml:"id,attr"`
	OffsetX    float64    `xml:"offsetx,attr"`
	OffsetY    float64    `xml:"offsety,attr"`
	ParallaxX  float64    `xml:"parallaxx,attr"`
	ParallaxY  float64    `xml:"parallaxy,attr"`
	Opacity    float64    `xml:"opacity,attr"`
	Visible    bool       `xml:"visible,attr"`
}

func defaultLayerBase() LayerBase {
	return LayerBase{ParallaxX: 1, ParallaxY: 1, Opacity: 1, Visible: true}
}

func (b LayerBase) validate() error {
	var errList []error
	if b.ID < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("id", b.ID, errMustNotBeNeg))
	}
	errList = append(errList, errs.CheckRange("opacity", b.Opacity, 0, 1))
	errList = append(errList, prefixed("properties", b.Properties.Validate()))
	return errors.Join(errList...)
}

// Layers is a list of layers in draw order. Each entry holds one kind of layer.
type Layers []Layer

type Layer struct {
	Tiles   *TileLayer
	Objects *ObjectGroup
	Image   *ImageLayer
	Group   *Group
}

func (ls *Layers) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var (
		l   Layer
		err error
	)
	switch start.Name.Local {
	case "layer":
		l.Tiles = &TileLayer{}
		err = d.DecodeElement(l.Tiles, &start)
	case "objectgroup":
		l.Objects = &ObjectGroup{}
		err = d.DecodeElement(l.Objects, &start)
	case "imagelayer":
		l.Image = &ImageLayer{}
		err = d.DecodeElement(l.Image, &start)
	case "group":
		l.Group = &Group{}
		err = d.DecodeElement(l.Group, &start)
	default:
		return d.Skip()
	}
	if err != nil {
		return err
	}
	*ls = append(*ls, l)
	return nil
}

// All walks every layer, going into groups after the group itself.
func (ls Layers) All() iter.Seq[Layer] {
	return func(yield func(Layer) bool) {
		ls.walk(yield)
	}
}

func (ls Layers) walk(yield func(Layer) bool) bool {
	for _, l := range ls {
		if !yield(l) {
			return false
		}
		if l.Group != nil && !l.Group.Layers.walk(yield) {
			return false
		}
	}
	return true
}

func (ls Layers) Validate() error {
	var errList []error
	for i, l := range ls {
		errList = append(errList, prefixed(fmt.Sprintf("layer[%d]", i), l.Validate()))
	}
	return errors.Join(errList...)
}

// Base is the shared part of whichever layer this is.
func (l Layer) Base() *LayerBase {
	switch {
	case l.Tiles != nil:
		return &l.Tiles.LayerBase
	case l.Objects != nil:
		return &l.Objects.LayerBase
	case l.Image != nil:
		return &l.Image.LayerBase
	case l.Group != nil:
		return &l.Group.LayerBase
	default:
		return nil
	}
}

func (l Layer) Validate() error {
	switch {
	case l.Tiles != nil:
		return l.Tiles.Validate()
	case l.Objects != nil:
		return l.Objects.Validate()
	case l.Image != nil:
		return l.Image.Validate()
	case l.Group != nil:
		return l.Group.Validate()
	default:
		return errs.Argument("layer", errMustNotBeEmpty)
	}
}

type BlendMode string

const (
	BlendNormal     BlendMode = "normal"
	BlendAdd        BlendMode = "add"
	BlendMultiply   BlendMode = "multiply"
	BlendScreen     BlendMode = "screen"
	BlendOverlay    BlendMode = "overlay"
	BlendDarken     BlendMode = "darken"
	BlendLighten    BlendMode = "lighten"
	BlendColorDodge BlendMode = "color-dodge"
	BlendColorBurn  BlendMode = "color-burn"
	BlendHardLight  BlendMode = "hard-light"
	BlendSoftLight  BlendMode = "soft-light"
	BlendDifference BlendMode = "difference"
	BlendExclusion  BlendMode = "exclusion"
)

func (m BlendMode) IsValid() bool {
	switch m {
	case BlendNormal, BlendAdd, BlendMultiply, BlendScreen, BlendOverlay, BlendDarken, BlendLighten,
		BlendColorDodge, BlendColorBurn, BlendHardLight, BlendSoftLight, BlendDifference, BlendExclusion:
		return true
	default:
		return false
	}
}

type TileLayer struct {
	Mode BlendMode `xml:"mode,attr"`
	Data LayerData `xml:"data"`
	LayerBase
	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`
}

func (l *TileLayer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain TileLayer
	*l = TileLayer{LayerBase: defaultLayerBase(), Mode: BlendNormal}
	return d.DecodeElement((*plain)(l), &start)
}

func (l TileLayer) Validate() error {
	errList := []error{l.validate()}
	if !l.Mode.IsValid() {
		errList = append(errList, errs.Argument("mode", fmt.Sprintf("unknown blend mode %q", l.Mode)))
	}
	if l.Width <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("width", l.Width, errMustBePositive))
	}
	if l.Height <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("height", l.Height, errMustBePositive))
	}
	errList = append(errList, l.Data.Validate())
	return errors.Join(errList...)
}

type DrawOrder string

const (
	DrawTopDown DrawOrder = "topdown"
	DrawIndex   DrawOrder = "index"
)

func (o DrawOrder) IsValid() bool {
	return o == DrawTopDown || o == DrawIndex
}

type ObjectGroup struct {
	DrawOrder DrawOrder `xml:"draworder,attr"`
	Objects   []Object  `xml:"object"`
	LayerBase
}

func (g *ObjectGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain ObjectGroup
	*g = ObjectGroup{LayerBase: defaultLayerBase(), DrawOrder: DrawTopDown}
	return d.DecodeElement((*plain)(g), &start)
}

func (g ObjectGroup) Validate() error {
	errList := []error{g.validate()}
	if !g.DrawOrder.IsValid() {
		errList = append(errList, errs.Argument("draworder", fmt.Sprintf("unknown draw order %q", g.DrawOrder)))
	}
	for i, o := range g.Objects {
		errList = append(errList, prefixed(fmt.Sprintf("object[%d]", i), o.Validate()))
	}
	return errors.Join(errList...)
}

type ImageLayer struct {
	Image *Image `xml:"image"`
	LayerBase
	RepeatX bool `xml:"repeatx,attr"`
	RepeatY bool `xml:"repeaty,attr"`
}

func (l *ImageLayer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain ImageLayer
	*l = ImageLayer{LayerBase: defaultLayerBase()}
	return d.DecodeElement((*plain)(l), &start)
}

// Validate allows a missing image, since Tiled saves empty image layers.
func (l ImageLayer) Validate() error {
	errList := []error{l.validate()}
	if l.Image != nil {
		errList = append(errList, prefixed("image", l.Image.Validate()))
	}
	return errors.Join(errList...)
}

type Group struct {
	Layers Layers `xml:",any"`
	LayerBase
}

func (g *Group) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Group
	*g = Group{LayerBase: defaultLayerBase()}
	return d.DecodeElement((*plain)(g), &start)
}

func (g Group) Validate() error {
	return errors.Join(g.validate(), g.Layers.Validate())
}
