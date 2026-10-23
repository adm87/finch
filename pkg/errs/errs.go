package errs

import (
	"cmp"
	"errors"
	"fmt"
)

var (
	ErrArgument           = errors.New("argument error")
	ErrArgumentNil        = errors.New("argument nil")
	ErrArgumentOutOfRange = errors.New("argument out of range")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrDisposed           = errors.New("object disposed")
	ErrNotSupported       = errors.New("not supported")
	ErrNotImplemented     = errors.New("not implemented")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrFormat             = errors.New("format error")
)

func withInner(msg string, inner error) string {
	if inner == nil {
		return msg
	}
	return msg + ": " + inner.Error()
}

func keyed(kind, key, suffix string) string {
	switch {
	case kind != "" && key != "":
		return fmt.Sprintf("%s %q %s", kind, key, suffix)
	case key != "":
		return fmt.Sprintf("key %q %s", key, suffix)
	case kind != "":
		return kind + " " + suffix
	default:
		return suffix
	}
}

// ArgumentError reports an invalid argument. errors.As with *ArgumentError
// also matches ArgumentNilError and ArgumentOutOfRangeError.
type ArgumentError struct {
	Inner   error
	Param   string
	Message string
}

func Argument(param, message string) *ArgumentError {
	return &ArgumentError{Param: param, Message: message}
}

func (e *ArgumentError) Wrap(err error) *ArgumentError {
	e.Inner = err
	return e
}

func (e *ArgumentError) Error() string {
	msg := "invalid argument"
	if e.Param != "" {
		msg = fmt.Sprintf("invalid argument %q", e.Param)
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return withInner(msg, e.Inner)
}

func (e *ArgumentError) Unwrap() error        { return e.Inner }
func (e *ArgumentError) Is(target error) bool { return target == ErrArgument }

// ArgumentNilError reports that a required parameter was nil.
type ArgumentNilError struct {
	Param string
}

func ArgumentNil(param string) *ArgumentNilError {
	return &ArgumentNilError{Param: param}
}

func (e *ArgumentNilError) Error() string {
	return fmt.Sprintf("argument %q cannot be nil", e.Param)
}

func (e *ArgumentNilError) Is(target error) bool {
	return target == ErrArgumentNil || target == ErrArgument
}

func (e *ArgumentNilError) As(target any) bool {
	if t, ok := target.(**ArgumentError); ok {
		*t = &ArgumentError{Param: e.Param, Message: "cannot be nil"}
		return true
	}
	return false
}

// ArgumentOutOfRangeError reports an argument outside its allowed range. Value holds the argument.
type ArgumentOutOfRangeError struct {
	Value   any
	Param   string
	Message string
}

func ArgumentOutOfRange(param string, value any, message string) *ArgumentOutOfRangeError {
	return &ArgumentOutOfRangeError{Param: param, Value: value, Message: message}
}

func (e *ArgumentOutOfRangeError) Error() string {
	msg := fmt.Sprintf("argument %q out of range (value %v)", e.Param, e.Value)
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return msg
}

func (e *ArgumentOutOfRangeError) Is(target error) bool {
	return target == ErrArgumentOutOfRange || target == ErrArgument
}

func (e *ArgumentOutOfRangeError) As(target any) bool {
	if t, ok := target.(**ArgumentError); ok {
		*t = &ArgumentError{Param: e.Param, Message: e.Error()}
		return true
	}
	return false
}

// CheckRange returns an ArgumentOutOfRangeError if value is outside [min, max].
func CheckRange[T cmp.Ordered](param string, value, min, max T) error {
	if value < min || value > max {
		return ArgumentOutOfRange(param, value, fmt.Sprintf("must be between %v and %v", min, max))
	}
	return nil
}

// InvalidOperationError reports a call that isn't valid in the object's current state.
type InvalidOperationError struct {
	Inner   error
	Message string
}

func InvalidOperation(message string) *InvalidOperationError {
	return &InvalidOperationError{Message: message}
}

func (e *InvalidOperationError) Wrap(err error) *InvalidOperationError {
	e.Inner = err
	return e
}

func (e *InvalidOperationError) Error() string {
	msg := "invalid operation"
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return withInner(msg, e.Inner)
}

func (e *InvalidOperationError) Unwrap() error        { return e.Inner }
func (e *InvalidOperationError) Is(target error) bool { return target == ErrInvalidOperation }

// DisposedError reports use of an object after it was closed or released.
// It also matches ErrInvalidOperation and *InvalidOperationError.
type DisposedError struct {
	Object string
}

func Disposed(object string) *DisposedError {
	return &DisposedError{Object: object}
}

func (e *DisposedError) Error() string {
	return fmt.Sprintf("cannot access disposed object %q", e.Object)
}

func (e *DisposedError) Is(target error) bool {
	return target == ErrDisposed || target == ErrInvalidOperation
}

func (e *DisposedError) As(target any) bool {
	if t, ok := target.(**InvalidOperationError); ok {
		*t = &InvalidOperationError{Message: e.Error()}
		return true
	}
	return false
}

// NotSupportedError reports an operation that will never be supported.
type NotSupportedError struct {
	Message string
}

func NotSupported(message string) *NotSupportedError {
	return &NotSupportedError{Message: message}
}

func (e *NotSupportedError) Error() string {
	if e.Message == "" {
		return "operation not supported"
	}
	return "operation not supported: " + e.Message
}

func (e *NotSupportedError) Is(target error) bool { return target == ErrNotSupported }

// NotImplementedError reports an operation that isn't written yet.
type NotImplementedError struct {
	Message string
}

func NotImplemented(message string) *NotImplementedError {
	return &NotImplementedError{Message: message}
}

func (e *NotImplementedError) Error() string {
	if e.Message == "" {
		return "not implemented"
	}
	return "not implemented: " + e.Message
}

func (e *NotImplementedError) Is(target error) bool { return target == ErrNotImplemented }

// NotFoundError reports a failed lookup. Kind names what was looked up, such as "image".
type NotFoundError struct {
	Inner error
	Kind  string
	Key   string
}

func NotFound(kind, key string) *NotFoundError {
	return &NotFoundError{Kind: kind, Key: key}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func (e *NotFoundError) Wrap(err error) *NotFoundError {
	e.Inner = err
	return e
}

func (e *NotFoundError) Error() string {
	return withInner(keyed(e.Kind, e.Key, "not found"), e.Inner)
}

func (e *NotFoundError) Unwrap() error        { return e.Inner }
func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }

// AlreadyExistsError reports a key that is already taken.
type AlreadyExistsError struct {
	Inner error
	Kind  string
	Key   string
}

func AlreadyExists(kind, key string) *AlreadyExistsError {
	return &AlreadyExistsError{Kind: kind, Key: key}
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

func (e *AlreadyExistsError) Wrap(err error) *AlreadyExistsError {
	e.Inner = err
	return e
}

func (e *AlreadyExistsError) Error() string {
	return withInner(keyed(e.Kind, e.Key, "already exists"), e.Inner)
}

func (e *AlreadyExistsError) Unwrap() error        { return e.Inner }
func (e *AlreadyExistsError) Is(target error) bool { return target == ErrAlreadyExists }

// FormatError reports input that couldn't be parsed. Input names the file or text.
type FormatError struct {
	Inner   error
	Input   string
	Message string
}

func Format(input, message string) *FormatError {
	return &FormatError{Input: input, Message: message}
}

func (e *FormatError) Wrap(err error) *FormatError {
	e.Inner = err
	return e
}

func (e *FormatError) Error() string {
	msg := fmt.Sprintf("invalid format %q", e.Input)
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return withInner(msg, e.Inner)
}

func (e *FormatError) Unwrap() error        { return e.Inner }
func (e *FormatError) Is(target error) bool { return target == ErrFormat }
