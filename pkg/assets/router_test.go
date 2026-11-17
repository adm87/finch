package assets_test

import (
	"errors"
	"testing"

	"github.com/adm87/finch/pkg/assets"
)

func newRouter(t *testing.T, c *assets.Cache[string]) *assets.Router {
	t.Helper()
	var r assets.Router
	if err := assets.Route(&r, c, ".txt"); err != nil {
		t.Fatalf("Route(.txt) = %v, want nil", err)
	}
	return &r
}

func TestRouteConflict(t *testing.T) {
	c, _ := newCache()
	r := newRouter(t, c)

	if err := assets.Route(r, c, ".txt"); !errors.Is(err, assets.ErrRouteConflict) {
		t.Errorf("Route(.txt) again = %v, want %v", err, assets.ErrRouteConflict)
	}
}

func TestRoutePanicsWithoutDot(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Route(txt) did not panic")
		}
	}()
	c, _ := newCache()
	var r assets.Router
	_ = assets.Route(&r, c, "txt")
}

func TestRouterLoadAndUnload(t *testing.T) {
	c, _ := newCache()
	r := newRouter(t, c)

	b, err := r.Load("a.txt", "b.txt", "a.txt")
	if err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}
	if got := c.Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}

	if err := b.Unload(); err != nil {
		t.Fatalf("Unload() = %v, want nil", err)
	}
	if err := b.Unload(); err != nil {
		t.Fatalf("second Unload() = %v, want nil", err)
	}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}

func TestRouterLoadFailsAllOrNothing(t *testing.T) {
	tests := []struct {
		want  error
		name  string
		paths []string
	}{
		{name: "missing file", paths: []string{"a.txt", "missing.txt"}},
		{name: "no route", paths: []string{"a.txt", "b.png"}, want: assets.ErrNoRoute},
		{name: "no extension", paths: []string{"a.txt", "readme"}, want: assets.ErrNoRoute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newCache()
			r := newRouter(t, c)

			b, err := r.Load(tt.paths...)
			if err == nil || (tt.want != nil && !errors.Is(err, tt.want)) {
				t.Errorf("Load(%v) = %v, want %v", tt.paths, err, tt.want)
			}
			if b != nil {
				t.Errorf("Load(%v) returned a bundle, want nil", tt.paths)
			}
			if got := c.Len(); got != 0 {
				t.Errorf("Len() = %d, want 0", got)
			}
		})
	}
}

func TestBundleUnloadKeepsSharedPaths(t *testing.T) {
	c, _ := newCache()
	r := newRouter(t, c)

	a, _ := r.Load("a.txt", "b.txt")
	_, _ = r.Load("a.txt")
	_ = a.Unload()

	if _, ok := c.Lookup("a.txt"); !ok {
		t.Error("Lookup(a.txt) = false, want true")
	}
	if _, ok := c.Lookup("b.txt"); ok {
		t.Error("Lookup(b.txt) = true, want false")
	}
}
