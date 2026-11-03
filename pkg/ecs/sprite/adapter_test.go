package sprite_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

type images []*ebiten.Image

func (i images) Image(handle uint64) *ebiten.Image {
	if handle >= uint64(len(i)) {
		return nil
	}
	return i[handle]
}

func newSprite(w donburi.World, handle uint64, x, y float64) *donburi.Entry {
	e := w.Entry(w.Create(render.Component, sprite.Component, transform.Component))
	sprite.Component.Get(e).SetHandle(handle)
	transform.Component.Get(e).SetPosition(x, y)
	return e
}

func TestAdapterAppendJobs(t *testing.T) {
	w := donburi.NewWorld()
	a := sprite.NewAdapter(images{nil, ebiten.NewImage(2, 2)})
	view := &render.View{Bounds: geom.Rect{Width: 10, Height: 10}}

	centred := newSprite(w, 1, -0.5, 5)
	sprite.Component.Get(centred).SetAnchor(0.5, 0.5)

	noTransform := w.Entry(w.Create(render.Component, sprite.Component))
	sprite.Component.Get(noTransform).SetHandle(1)

	tests := []struct {
		name  string
		entry *donburi.Entry
		want  int
	}{
		{"inside view", newSprite(w, 1, 4, 4), 1},
		{"outside view", newSprite(w, 1, 20, 4), 0},
		{"touching edge", newSprite(w, 1, 10, 4), 0},
		{"anchor overlaps edge", centred, 1},
		{"no transform", noTransform, 0},
		{"no sprite", w.Entry(w.Create(render.Component, transform.Component)), 0},
		{"no image", newSprite(w, 0, 4, 4), 0},
	}
	transform.Update(w)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(a.AppendJobs(tt.entry, view, nil)); got != tt.want {
				t.Errorf("AppendJobs returned %d jobs, want %d", got, tt.want)
			}
		})
	}
}
