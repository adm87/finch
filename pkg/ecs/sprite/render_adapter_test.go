package sprite_test

import (
	"testing"

	"github.com/adm87/finch/pkg/ecs/sprite"
)

func TestRenderAdapterPanicsOnNilTextures(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("RenderAdapter(nil) did not panic")
		}
	}()
	sprite.RenderAdapter(nil)
}
