package assets_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/adm87/finch/pkg/assets"
)

func TestContentOpensFromEachMount(t *testing.T) {
	var c assets.Content
	_ = c.Mount(fstest.MapFS{"images/a.png": {Data: []byte("a")}})
	_ = c.Mount(fstest.MapFS{"audio/b.ogg": {Data: []byte("b")}})

	for path, want := range map[string]string{"images/a.png": "a", "audio/b.ogg": "b"} {
		got, err := fs.ReadFile(&c, path)
		if err != nil || string(got) != want {
			t.Errorf("ReadFile(%s) = %q, %v, want %q, nil", path, got, err, want)
		}
	}
}

func TestContentOpenFails(t *testing.T) {
	var c assets.Content
	_ = c.Mount(fstest.MapFS{"a.txt": {Data: []byte("a")}})

	tests := []struct {
		want error
		name string
		path string
	}{
		{name: "missing file", path: "missing.txt", want: fs.ErrNotExist},
		{name: "invalid path", path: "../a.txt", want: fs.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := c.Open(tt.path); !errors.Is(err, tt.want) {
				t.Errorf("Open(%s) = %v, want %v", tt.path, err, tt.want)
			}
		})
	}
}

func TestContentMountConflictMountsNothing(t *testing.T) {
	var c assets.Content
	_ = c.Mount(fstest.MapFS{"a.txt": {Data: []byte("a")}})

	err := c.Mount(fstest.MapFS{
		"a.txt": {Data: []byte("other")},
		"b.txt": {Data: []byte("b")},
	})
	if !errors.Is(err, assets.ErrMountConflict) {
		t.Errorf("Mount() = %v, want %v", err, assets.ErrMountConflict)
	}
	if _, err := c.Open("b.txt"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Open(b.txt) = %v, want %v", err, fs.ErrNotExist)
	}
	if got, _ := fs.ReadFile(&c, "a.txt"); string(got) != "a" {
		t.Errorf("ReadFile(a.txt) = %q, want %q", got, "a")
	}
}

func TestContentMountPanicsOnNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Mount(nil) did not panic")
		}
	}()
	var c assets.Content
	_ = c.Mount(nil)
}
