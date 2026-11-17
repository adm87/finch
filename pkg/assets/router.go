package assets

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

var (
	ErrNoRoute       = errors.New("assets: no route for extension")
	ErrRouteConflict = errors.New("assets: extension already routed")
)

type route struct {
	load   func(path string) error
	unload func(path string) error
}

// Router routes file extensions to a Cache. Use it to load bundles of assets that contain mixed types.
// Loading a bundle returns a Bundle handle that can then be used to unload all the assets in the bundle at once.
type Router struct {
	routes map[string]route
}

func Route[T any](r *Router, c *Cache[T], exts ...string) error {
	var conflicts []error

	for _, ext := range exts {
		if !strings.HasPrefix(ext, ".") {
			panic(fmt.Sprintf("assets: extension %q needs a leading dot", ext))
		}
		if _, ok := r.routes[ext]; ok {
			conflicts = append(conflicts, fmt.Errorf("route %q: %w", ext, ErrRouteConflict))
		}
	}

	if len(conflicts) > 0 {
		return errors.Join(conflicts...)
	}

	if r.routes == nil {
		r.routes = make(map[string]route)
	}

	for _, ext := range exts {
		r.routes[ext] = route{
			load: func(path string) error {
				_, err := c.Load(path)
				return err
			},
			unload: c.Unload,
		}
	}

	return nil
}

func (r *Router) Load(paths ...string) (*Bundle, error) {
	paths = distinct(paths)

	loaded := make([]string, 0, len(paths))

	var errs []error
	for _, p := range paths {
		rt, ok := r.routes[path.Ext(p)]

		if !ok {
			errs = append(errs, fmt.Errorf("load %q: %w", p, ErrNoRoute))
			continue
		}

		if err := rt.load(p); err != nil {
			errs = append(errs, err)
			continue
		}

		loaded = append(loaded, p)
	}

	b := &Bundle{router: r, paths: loaded}
	if len(errs) > 0 {
		errs = append(errs, b.Unload())
		return nil, errors.Join(errs...)
	}

	return b, nil
}

type Bundle struct {
	router *Router
	paths  []string
}

// Unload unloads every path the bundle loaded. Calling it again does nothing.
func (b *Bundle) Unload() error {
	if b == nil {
		return nil
	}
	var errs []error
	for _, p := range b.paths {
		if err := b.router.routes[path.Ext(p)].unload(p); err != nil {
			errs = append(errs, err)
		}
	}
	b.paths = nil
	return errors.Join(errs...)
}

func distinct(paths []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, path := range paths {
		if _, ok := seen[path]; !ok {
			seen[path] = struct{}{}
			result = append(result, path)
		}
	}

	return result
}
