package resource

import (
	"errors"
	"fmt"
)

// Request is a set of files to read from one source.
type Request struct {
	Source Source
	Names  []string
}

// Result is one file a Loader read, or the error that stopped it.
type Result[T any] struct {
	Value  T
	Err    error
	Source Source
	Name   string
}

// Loader reads queued files and decodes each into a T. It holds nothing once
// Load returns, so make one per batch and throw it away.
type Loader[T any] struct {
	res    *Resources
	decode func([]byte) (T, error)
	queue  []Request
}

func NewLoader[T any](res *Resources, decode func([]byte) (T, error)) *Loader[T] {
	return &Loader[T]{res: res, decode: decode}
}

func (l *Loader[T]) Queue(src Source, names ...string) {
	l.queue = append(l.queue, Request{Source: src, Names: names})
}

// Load reads and decodes everything queued, in order. A file queued twice is
// read once. Every file gets a result, so one failure doesn't stop the rest.
func (l *Loader[T]) Load() []Result[T] {
	seen := make(map[string]struct{})
	var results []Result[T]
	for _, req := range l.queue {
		for _, name := range req.Names {
			key := Key(req.Source, name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			results = append(results, l.load(req.Source, name))
		}
	}
	l.queue = nil
	return results
}

func (l *Loader[T]) load(src Source, name string) Result[T] {
	r := Result[T]{Source: src, Name: name}
	raw, err := l.res.ReadFile(src, name)
	if err != nil {
		r.Err = err
		return r
	}
	if r.Value, err = l.decode(raw); err != nil {
		r.Err = fmt.Errorf("decode %s: %w", Key(src, name), err)
	}
	return r
}

// Bytes decodes a file as its raw bytes.
func Bytes(raw []byte) ([]byte, error) {
	return raw, nil
}

// Err joins the errors in results, or returns nil if every file loaded.
func Err[T any](results []Result[T]) error {
	var errList []error
	for _, r := range results {
		if r.Err != nil {
			errList = append(errList, r.Err)
		}
	}
	return errors.Join(errList...)
}
