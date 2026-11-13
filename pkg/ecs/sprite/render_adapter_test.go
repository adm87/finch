package sprite_test

import (
	"testing"

	"github.com/adm87/finch/pkg/ecs/sprite"
)

func TestNewRenderAdapterPanicsOnNilImages(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewRenderAdapter(nil) did not panic")
		}
	}()
	sprite.NewRenderAdapter(nil)
}
