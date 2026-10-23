package slotmap

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/errs"
)

// HandleError converts a SlotMap error for a packed handle into an errs error.
func HandleError(op, kind string, handle uint64, err error) error {
	key := strconv.FormatUint(handle, 10)
	switch {
	case errors.Is(err, ErrInvalidKey):
		return errs.Argument("handle", fmt.Sprintf("malformed %s handle %s", kind, key)).Wrap(err)
	case errors.Is(err, ErrNotFound):
		return errs.NotFound(kind, key).Wrap(err)
	default:
		return fmt.Errorf("%s %s %s: %w", op, kind, key, err)
	}
}
