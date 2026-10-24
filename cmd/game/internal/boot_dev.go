//go:build finch_dev

package internal

import (
	"github.com/adm87/finch/pkg/logger"
)

var loggerOptions = logger.Options{
	Level:      logger.DebugLevel,
	WithCaller: true,
}
