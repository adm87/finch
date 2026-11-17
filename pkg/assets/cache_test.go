package assets_test

import (
	"errors"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/adm87/finch/pkg/assets"
)

// newCache loads a.txt and b.txt as strings and records what it releases.
func newCache() (*assets.Cache[string], *[]string) {
	var released []string
	fsys := fstest.MapFS{
		"a.txt": {Data: []byte("a")},
		"b.txt": {Data: []byte("b")},
	}
	importer := assets.ImporterFunc[string](func(fsys fs.FS, path string) (string, error) {
		data, err := fs.ReadFile(fsys, path)
		return string(data), err
	})
	releaser := assets.ReleaserFunc[string](func(value string) {
		released = append(released, value)
	})
	return assets.NewCache(fsys, importer, releaser), &released
}

func TestCacheLoadTwiceSharesOneCopy(t *testing.T) {
	c, _ := newCache()

	h1, err := c.Load("a.txt")
	if err != nil {
		t.Fatalf("Load(a.txt) = %v, want nil", err)
	}
	h2, _ := c.Load("a.txt")

	if h1 != h2 {
		t.Errorf("Load(a.txt) twice = %v and %v, want the same handle", h1, h2)
	}
	if got, _ := c.Get(h1); got != "a" {
		t.Errorf("Get(h) = %q, want %q", got, "a")
	}
	if got := c.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestCacheLoadMissingFile(t *testing.T) {
	c, _ := newCache()

	if _, err := c.Load("missing.txt"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load(missing.txt) = %v, want %v", err, fs.ErrNotExist)
	}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}

func TestCacheUnloadReleasesAtZero(t *testing.T) {
	c, released := newCache()

	h, _ := c.Load("a.txt")
	_, _ = c.Load("a.txt")

	_ = c.Unload("a.txt")
	if len(*released) != 0 {
		t.Errorf("released after one Unload = %v, want nothing", *released)
	}

	_ = c.Unload("a.txt")
	if !slices.Equal(*released, []string{"a"}) {
		t.Errorf("released after two Unloads = %v, want [a]", *released)
	}
	if _, ok := c.Get(h); ok {
		t.Error("Get(h) after release = true, want false")
	}
}

func TestCacheUnloadNotLoaded(t *testing.T) {
	c, _ := newCache()

	if err := c.Unload("a.txt"); !errors.Is(err, assets.ErrNotLoaded) {
		t.Errorf("Unload(a.txt) = %v, want %v", err, assets.ErrNotLoaded)
	}
}

func TestCacheClearReleasesEverything(t *testing.T) {
	c, released := newCache()

	_, _ = c.Load("a.txt")
	_, _ = c.Load("a.txt")
	_, _ = c.Load("b.txt")
	c.Clear()

	if got := c.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
	if len(*released) != 2 {
		t.Errorf("released = %v, want a and b", *released)
	}
}
