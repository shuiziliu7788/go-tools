package log

import (
	"log/slog"
	"os"
	"sync/atomic"
)

var root atomic.Value

func init() {
	root.Store(&Logger{
		inner: slog.Default(),
	})
}

func SetDefault(l *Logger) {
	root.Store(l)
	slog.SetDefault(l.inner)
}

func Default() *Logger {
	return root.Load().(*Logger)
}

func Debug(msg string, args ...any) {
	Default().Log(slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	Default().Log(slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Log(slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	Default().Log(slog.LevelError, msg, args...)
}

func Fatal(msg string, ctx ...interface{}) {
	Default().Log(slog.LevelError, msg, ctx...)
	os.Exit(1)
}

func New(ctx ...interface{}) *Logger {
	return Default().With(ctx...)
}
