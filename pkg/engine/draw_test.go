package engine_test

import (
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

func TestDrawPipelineWrapsInOrder(t *testing.T) {
	var calls []string

	record := func(name string) engine.DrawPass {
		return func(inner engine.Drawer) engine.Drawer {
			return engine.DrawFunc(func(target *ebiten.Image) {
				calls = append(calls, name+" before")
				inner.Draw(target)
				calls = append(calls, name+" after")
			})
		}
	}
	base := engine.DrawFunc(func(*ebiten.Image) {
		calls = append(calls, "base")
	})

	engine.DrawPipeline(base, record("a"), record("b")).Draw(nil)

	want := []string{"b before", "a before", "base", "a after", "b after"}
	if !slices.Equal(calls, want) {
		t.Errorf("calls = %q, want %q", calls, want)
	}
}

func TestDrawPipelineWithoutPasses(t *testing.T) {
	drawn := false
	base := engine.DrawFunc(func(*ebiten.Image) { drawn = true })

	engine.DrawPipeline(base).Draw(nil)

	if !drawn {
		t.Error("base drawer was not called")
	}
}
