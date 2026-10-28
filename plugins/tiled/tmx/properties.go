package tmx

import (
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/adm87/finch/pkg/errs"
)

// Color is a Tiled color, written as #AARRGGBB or #RRGGBB.
type Color color.NRGBA

func (c *Color) UnmarshalText(text []byte) error {
	s := strings.TrimPrefix(string(text), "#")
	switch len(s) {
	case 6:
		s = "ff" + s
	case 8:
	default:
		return errs.Format(string(text), "want #RRGGBB or #AARRGGBB")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return errs.Format(string(text), "invalid color").Wrap(err)
	}
	*c = Color{A: b[0], R: b[1], G: b[2], B: b[3]}
	return nil
}

func (c *Color) RGBA() (r, g, b, a uint32) {
	return color.NRGBA(*c).RGBA()
}

type PropertyType string

const (
	PropertyString PropertyType = "string"
	PropertyInt    PropertyType = "int"
	PropertyFloat  PropertyType = "float"
	PropertyBool   PropertyType = "bool"
	PropertyColor  PropertyType = "color"
	PropertyFile   PropertyType = "file"
	PropertyObject PropertyType = "object"
	PropertyClass  PropertyType = "class"
)

func (t PropertyType) IsValid() bool {
	switch t {
	case PropertyString, PropertyInt, PropertyFloat, PropertyBool, PropertyColor, PropertyFile, PropertyObject, PropertyClass:
		return true
	default:
		return false
	}
}

// Property is a custom property. Class properties keep their set members in
// Members.
type Property struct {
	Name         string       `xml:"name,attr"`
	Type         PropertyType `xml:"type,attr"`
	PropertyType string       `xml:"propertytype,attr"`
	Value        string       `xml:"value,attr"`
	Members      Properties   `xml:"properties>property"`
}

func (p *Property) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Property
	*p = Property{Type: PropertyString}

	// Multi-line strings are written as the element's text.
	aux := struct {
		*plain
		Text string `xml:",chardata"`
	}{plain: (*plain)(p)}
	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}
	if p.Value == "" && p.Type == PropertyString {
		p.Value = aux.Text
	}
	return nil
}

func (p *Property) Int() (int, error) {
	if p.Value == "" {
		return 0, nil
	}
	return strconv.Atoi(p.Value)
}

func (p *Property) Float() (float64, error) {
	if p.Value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(p.Value, 64)
}

func (p *Property) Bool() (bool, error) {
	if p.Value == "" {
		return false, nil
	}
	return strconv.ParseBool(p.Value)
}

func (p *Property) Color() (Color, error) {
	var c Color
	if p.Value == "" {
		return c, nil
	}
	err := c.UnmarshalText([]byte(p.Value))
	return c, err
}

// ObjectID is the referenced object's ID, or 0 for none.
func (p *Property) ObjectID() (int, error) {
	return p.Int()
}

func (p *Property) Validate() error {
	var err error
	switch p.Type {
	case PropertyString, PropertyFile:
	case PropertyInt, PropertyObject:
		_, err = p.Int()
	case PropertyFloat:
		_, err = p.Float()
	case PropertyBool:
		_, err = p.Bool()
	case PropertyColor:
		_, err = p.Color()
	case PropertyClass:
		return p.Members.Validate()
	default:
		return errs.Argument("type", fmt.Sprintf("unknown property type %q", p.Type))
	}
	if err != nil {
		return errs.Argument("value", fmt.Sprintf("%q is not a valid %s", p.Value, p.Type)).Wrap(err)
	}
	return nil
}

type Properties []Property

func (ps Properties) Get(name string) (Property, bool) {
	for _, p := range ps {
		if p.Name == name {
			return p, true
		}
	}
	return Property{}, false
}

func (ps Properties) Validate() error {
	var errList []error
	for _, p := range ps {
		errList = append(errList, prefixed(fmt.Sprintf("property %q", p.Name), p.Validate()))
	}
	return errors.Join(errList...)
}
