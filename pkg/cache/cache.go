package cache

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/types/errs"
	"github.com/adm87/finch/pkg/types/slotmap"
)

func cacheError(op, store string, handle uint64, err error) error {
	key := strconv.FormatUint(handle, 10)
	switch {
	case errors.Is(err, slotmap.ErrInvalidKey):
		return errs.Argument("handle", fmt.Sprintf("malformed %s handle %s", store, key)).Wrap(err)
	case errors.Is(err, slotmap.ErrNotFound):
		return errs.NotFound(store, key).Wrap(err)
	default:
		return fmt.Errorf("%s %s %s: %w", op, store, key, err)
	}
}
