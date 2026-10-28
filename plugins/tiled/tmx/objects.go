package tmx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/adm87/finch/pkg/errs"
)

type ObjectShape string

const (
	ShapeRectangle ObjectShape = "rectangle"
	ShapeEllipse   ObjectShape = "ellipse"
	ShapeCapsule   ObjectShape = "capsule"
	ShapePoint     ObjectShape = "point"
	ShapePolygon   ObjectShape = "polygon"
	ShapePolyline  ObjectShape = "polyline"
	ShapeText      ObjectShape = "text"
	ShapeTile      ObjectShape = "tile"
)

// Object is a shape, tile or text placed on an object layer. Points and
// polygons have no size.
type Object struct {
	Ellipse    *struct{}  `xml:"ellipse"`
	Capsule    *struct{}  `xml:"capsule"`
	Point      *struct{}  `xml:"point"`
	Polygon    *Poly      `xml:"polygon"`
	Polyline   *Poly      `xml:"polyline"`
	Text       *Text      `xml:"text"`
	Name       string     `xml:"name,attr"`
	Type       string     `xml:"type,attr"` // class before Tiled 1.9
	Class      string     `xml:"class,attr"`
	Template   string     `xml:"template,attr"` // .tx file, relative to the map
	Properties Properties `xml:"properties>property"`
	ID         int        `xml:"id,attr"`
	X          float64    `xml:"x,attr"`
	Y          float64    `xml:"y,attr"`
	Width      float64    `xml:"width,attr"`
	Height     float64    `xml:"height,attr"`
	Rotation   float64    `xml:"rotation,attr"` // degrees clockwise
	Opacity    float64    `xml:"opacity,attr"`
	GID        GID        `xml:"gid,attr"`
	Visible    bool       `xml:"visible,attr"`
}

func (o *Object) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Object
	*o = Object{Opacity: 1, Visible: true}
	return d.DecodeElement((*plain)(o), &start)
}

// ClassName is the class, whichever attribute Tiled saved it in.
func (o Object) ClassName() string {
	if o.Class != "" {
		return o.Class
	}
	return o.Type
}

func (o Object) Shape() ObjectShape {
	switch {
	case o.Ellipse != nil:
		return ShapeEllipse
	case o.Capsule != nil:
		return ShapeCapsule
	case o.Point != nil:
		return ShapePoint
	case o.Polygon != nil:
		return ShapePolygon
	case o.Polyline != nil:
		return ShapePolyline
	case o.Text != nil:
		return ShapeText
	case o.GID.ID() != 0:
		return ShapeTile
	default:
		return ShapeRectangle
	}
}

func (o Object) Validate() error {
	var errList []error
	if o.ID < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("id", o.ID, errMustNotBeNeg))
	}
	if o.Width < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("width", o.Width, errMustNotBeNeg))
	}
	if o.Height < 0 {
		errList = append(errList, errs.ArgumentOutOfRange("height", o.Height, errMustNotBeNeg))
	}
	errList = append(errList, errs.CheckRange("opacity", o.Opacity, 0, 1))
	if o.Polygon != nil && len(o.Polygon.Points) == 0 {
		errList = append(errList, errs.Argument("polygon", errMustNotBeEmpty))
	}
	if o.Polyline != nil && len(o.Polyline.Points) == 0 {
		errList = append(errList, errs.Argument("polyline", errMustNotBeEmpty))
	}
	if o.Text != nil {
		errList = append(errList, prefixed("text", o.Text.Validate()))
	}
	errList = append(errList, prefixed("properties", o.Properties.Validate()))
	return errors.Join(errList...)
}

type Point struct {
	X, Y float64
}

// Points are relative to the object's position.
type Points []Point

func (ps *Points) UnmarshalText(text []byte) error {
	fields := strings.Fields(string(text))
	out := make(Points, len(fields))
	for i, f := range fields {
		xs, ys, ok := strings.Cut(f, ",")
		if !ok {
			return errs.Format(f, "want x,y")
		}
		x, err := strconv.ParseFloat(xs, 64)
		if err != nil {
			return errs.Format(f, "invalid x").Wrap(err)
		}
		y, err := strconv.ParseFloat(ys, 64)
		if err != nil {
			return errs.Format(f, "invalid y").Wrap(err)
		}
		out[i] = Point{X: x, Y: y}
	}
	*ps = out
	return nil
}

type Poly struct {
	Points Points `xml:"points,attr"`
}

type HAlign string

const (
	HAlignLeft    HAlign = "left"
	HAlignCenter  HAlign = "center"
	HAlignRight   HAlign = "right"
	HAlignJustify HAlign = "justify"
)

func (a HAlign) IsValid() bool {
	switch a {
	case HAlignLeft, HAlignCenter, HAlignRight, HAlignJustify:
		return true
	default:
		return false
	}
}

type VAlign string

const (
	VAlignTop    VAlign = "top"
	VAlignCenter VAlign = "center"
	VAlignBottom VAlign = "bottom"
)

func (a VAlign) IsValid() bool {
	switch a {
	case VAlignTop, VAlignCenter, VAlignBottom:
		return true
	default:
		return false
	}
}

// Text is clipped to its object's bounds.
type Text struct {
	Content    string `xml:",chardata"`
	FontFamily string `xml:"fontfamily,attr"`
	HAlign     HAlign `xml:"halign,attr"`
	VAlign     VAlign `xml:"valign,attr"`
	Color      Color  `xml:"color,attr"`
	PixelSize  int    `xml:"pixelsize,attr"`
	Wrap       bool   `xml:"wrap,attr"`
	Bold       bool   `xml:"bold,attr"`
	Italic     bool   `xml:"italic,attr"`
	Underline  bool   `xml:"underline,attr"`
	Strikeout  bool   `xml:"strikeout,attr"`
	Kerning    bool   `xml:"kerning,attr"`
}

func (t *Text) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Text
	*t = Text{
		FontFamily: "sans-serif",
		HAlign:     HAlignLeft,
		VAlign:     VAlignTop,
		Color:      Color{A: 255},
		PixelSize:  16,
		Kerning:    true,
	}
	return d.DecodeElement((*plain)(t), &start)
}

func (t Text) Validate() error {
	var errList []error
	if !t.HAlign.IsValid() {
		errList = append(errList, errs.Argument("halign", fmt.Sprintf("unknown alignment %q", t.HAlign)))
	}
	if !t.VAlign.IsValid() {
		errList = append(errList, errs.Argument("valign", fmt.Sprintf("unknown alignment %q", t.VAlign)))
	}
	if t.PixelSize <= 0 {
		errList = append(errList, errs.ArgumentOutOfRange("pixelsize", t.PixelSize, errMustBePositive))
	}
	return errors.Join(errList...)
}
