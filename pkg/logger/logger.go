package logger

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zap.Field

type LogLevel = zapcore.Level

const (
	DebugLevel  = zapcore.DebugLevel
	InfoLevel   = zapcore.InfoLevel
	WarnLevel   = zapcore.WarnLevel
	ErrorLevel  = zapcore.ErrorLevel
	DPanicLevel = zapcore.DPanicLevel
	PanicLevel  = zapcore.PanicLevel
	FatalLevel  = zapcore.FatalLevel
)

type Options struct {
	Level      LogLevel
	WithCaller bool
}

type Logger interface {
	With(fields ...Field) Logger
	Named(name string) Logger
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Sync()
}

func NewLogger(out io.Writer, opts Options) Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.ConsoleSeparator = " "

	var zOpts []zap.Option
	if opts.WithCaller {
		zOpts = append(zOpts, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		cfg.CallerKey = ""
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), zapcore.AddSync(out), opts.Level)
	return &logger{zap: zap.New(core, zOpts...)}
}

type logger struct {
	zap *zap.Logger
}

func (l *logger) With(fields ...Field) Logger {
	return &logger{zap: l.zap.With(fields...)}
}

func (l *logger) Named(name string) Logger {
	return &logger{zap: l.zap.Named(name)}
}

func (l *logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *logger) Sync() {
	_ = l.zap.Sync()
}

func String(key, value string) Field {
	return zap.String(key, value)
}

func Strings(key string, value []string) Field {
	return zap.Strings(key, value)
}

func Int(key string, value int) Field {
	return zap.Int(key, value)
}

func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

func Float32(key string, value float32) Field {
	return zap.Float32(key, value)
}

func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

func ErrorField(key string, err error) Field {
	return zap.NamedError(key, err)
}

func Reflect(key string, value any) Field {
	return zap.Any(key, value)
}

var N Logger = &noopLogger{}

type noopLogger struct{}

func (n *noopLogger) With(fields ...Field) Logger {
	return n
}

func (n *noopLogger) Named(name string) Logger {
	return n
}

func (n *noopLogger) Info(msg string, fields ...Field) {
	// no-op
}

func (n *noopLogger) Error(msg string, fields ...Field) {
	// no-op
}

func (n *noopLogger) Fatal(msg string, fields ...Field) {
	// no-op
}

func (n *noopLogger) Debug(msg string, fields ...Field) {
	// no-op
}

func (n *noopLogger) Warn(msg string, fields ...Field) {
	// no-op
}

func (n *noopLogger) Sync() {
	// no-op
}
