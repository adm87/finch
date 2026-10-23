package aseprite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hexcolor"
)

// Atlas is an Aseprite sprite sheet export.
type Atlas struct {
	Frames Frames `json:"frames"`
	Meta   Meta   `json:"meta"`
}

// DropMissingTags removes frame tags whose frames were not exported and returns their names.
func (a *Atlas) DropMissingTags() []string {
	var dropped []string
	kept := a.Meta.FrameTags[:0]
	for _, tag := range a.Meta.FrameTags {
		if tag.From >= len(a.Frames) {
			dropped = append(dropped, tag.Name)
			continue
		}
		kept = append(kept, tag)
	}
	a.Meta.FrameTags = kept
	return dropped
}

// Validate checks that every frame tag has a known direction and a valid frame range.
func (a *Atlas) Validate() error {
	var errList []error
	for _, tag := range a.Meta.FrameTags {
		if tag.From < 0 || tag.To >= len(a.Frames) || tag.From > tag.To {
			msg := fmt.Sprintf("frame range %d..%d is outside 0..%d", tag.From, tag.To, len(a.Frames)-1)
			errList = append(errList, errs.Format(tag.Name, msg))
		}
		if !tag.Direction.IsValid() {
			msg := fmt.Sprintf("unknown animation direction %q", tag.Direction)
			errList = append(errList, errs.Format(tag.Name, msg))
		}
	}
	return errors.Join(errList...)
}

// Frames holds an atlas's frames in file order.
type Frames []Frame

// Rectangles returns the bounding rectangles of all frames in the atlas.
func (f Frames) Rectangles() []image.Rectangle {
	rects := make([]image.Rectangle, len(f))
	for i, frame := range f {
		rects[i] = frame.FrameRect.ToImageRectangle()
	}
	return rects
}

func (f Frames) Slice(from, to int) Frames {
	if from < 0 {
		from = 0
	}
	if to > len(f) {
		to = len(f)
	}
	return f[from:to]
}

func (f Frames) Len() int {
	return len(f)
}

// UnmarshalJSON decodes both the Array and Hash export layouts.
// Hash keys become each frame's FileName.
func (f *Frames) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) == 0 {
		return errs.Format("frames", "empty value")
	}

	switch trimmed[0] {
	case '[':
		var frames []Frame
		if err := json.Unmarshal(data, &frames); err != nil {
			return errs.Format("frames", "invalid array layout").Wrap(err)
		}
		*f = frames
		return nil
	case '{':
		frames, err := decodeFrameHash(data)
		if err != nil {
			return errs.Format("frames", "invalid hash layout").Wrap(err)
		}
		*f = frames
		return nil
	default:
		return errs.Format("frames", "expected an array or an object")
	}
}

func decodeFrameHash(data []byte) ([]Frame, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}

	var frames []Frame
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		name, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("expected frame name, got %v", tok)
		}

		var frame Frame
		if err := dec.Decode(&frame); err != nil {
			return nil, fmt.Errorf("frame %q: %w", name, err)
		}
		frame.FileName = name
		frames = append(frames, frame)
	}
	return frames, nil
}

type Frame struct {
	FileName         string    `json:"filename"`
	FrameRect        geom.Rect `json:"frame"`
	Rotated          bool      `json:"rotated"`
	Trimmed          bool      `json:"trimmed"`
	SpriteSourceSize geom.Rect `json:"spriteSourceSize"`
	SourceSize       geom.Rect `json:"sourceSize"`
	Duration         int       `json:"duration"`
}

type Meta struct {
	App       string    `json:"app"`
	Version   string    `json:"version"`
	Image     string    `json:"image"`
	Format    string    `json:"format"`
	FrameTags FrameTags `json:"frameTags"`
	Size      geom.Rect `json:"size"`
	Scale     float64   `json:"scale,string"`
}

type FrameTags []FrameTag

func (ft FrameTags) ByName(name string) *FrameTag {
	for i := range ft {
		if ft[i].Name == name {
			return &ft[i]
		}
	}
	return nil
}

type FrameTag struct {
	Name      string         `json:"name"`
	Direction Direction      `json:"direction"`
	From      int            `json:"from"`
	To        int            `json:"to"`
	Color     hexcolor.Color `json:"color"`
}

// Direction is the playback order of a frame tag.
type Direction string

const (
	DirectionForward         Direction = "forward"
	DirectionReverse         Direction = "reverse"
	DirectionPingPong        Direction = "pingpong"
	DirectionPingPongReverse Direction = "pingpong_reverse"
)

func (d Direction) IsValid() bool {
	switch d {
	case DirectionForward, DirectionReverse, DirectionPingPong, DirectionPingPongReverse:
		return true
	default:
		return false
	}
}

func (d Direction) String() string {
	return string(d)
}
