package errs

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"testing"
)

func TestArgumentNilHierarchy(t *testing.T) {
	err := fmt.Errorf("load sprite: %w", ArgumentNil("handle"))

	if !errors.Is(err, ErrArgumentNil) || !errors.Is(err, ErrArgument) {
		t.Fatal("expected ErrArgumentNil and ErrArgument")
	}
	if errors.Is(err, ErrArgumentOutOfRange) {
		t.Fatal("unexpected ErrArgumentOutOfRange")
	}

	var nilErr *ArgumentNilError
	if !errors.As(err, &nilErr) || nilErr.Param != "handle" {
		t.Fatal("expected *ArgumentNilError with Param handle")
	}

	var argErr *ArgumentError
	if !errors.As(err, &argErr) || argErr.Param != "handle" {
		t.Fatal("expected *ArgumentError with Param handle")
	}
}

func TestCheckRange(t *testing.T) {
	if err := CheckRange("count", 5, 0, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := CheckRange("count", 42, 0, 10)
	var oor *ArgumentOutOfRangeError
	if !errors.As(err, &oor) || oor.Value != 42 {
		t.Fatal("expected *ArgumentOutOfRangeError with Value 42")
	}
	if !errors.Is(err, ErrArgument) {
		t.Fatal("expected ErrArgument")
	}
	want := `argument "count" out of range (value 42): must be between 0 and 10`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestDisposedIsInvalidOperation(t *testing.T) {
	var err error = Disposed("ImageCache")

	if !errors.Is(err, ErrDisposed) || !errors.Is(err, ErrInvalidOperation) {
		t.Fatal("expected ErrDisposed and ErrInvalidOperation")
	}
	if _, ok := errors.AsType[*InvalidOperationError](err); !ok {
		t.Fatal("expected As *InvalidOperationError")
	}
}

func TestWithInnerKeepsTheChain(t *testing.T) {
	cause := fs.ErrNotExist
	cases := map[string]struct {
		err      error
		sentinel error
	}{
		"Argument":         {Argument("p", "bad").Wrap(cause), ErrArgument},
		"InvalidOperation": {InvalidOperation("read file").Wrap(cause), ErrInvalidOperation},
		"NotFound":         {NotFound("resource", "a.png").Wrap(cause), ErrNotFound},
		"AlreadyExists":    {AlreadyExists("file system", "assets").Wrap(cause), ErrAlreadyExists},
		"Format":           {Format("cfg.yaml", "bad yaml").Wrap(cause), ErrFormat},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			wrapped := fmt.Errorf("outer: %w", tc.err)
			if !errors.Is(wrapped, tc.sentinel) {
				t.Error("lost category sentinel")
			}
			if !errors.Is(wrapped, cause) {
				t.Error("lost inner cause")
			}
		})
	}
}

func TestWithInnerExternalType(t *testing.T) {
	_, perr := strconv.Atoi("12x")
	err := Format("12x", "expected integer").Wrap(perr)

	var numErr *strconv.NumError
	if !errors.Is(err, ErrFormat) || !errors.As(err, &numErr) {
		t.Fatal("expected ErrFormat and *strconv.NumError")
	}
}

func TestInnerIsOptional(t *testing.T) {
	want := `file system "assets" already exists`
	if got := AlreadyExists("file system", "assets").Error(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if errors.Unwrap(NotFound("image", "x")) != nil {
		t.Fatal("expected no inner error")
	}
}

func TestCheckRangeReturnsTrueNil(t *testing.T) {
	var err = CheckRange("x", 1, 0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
