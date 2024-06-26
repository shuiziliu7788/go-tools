package log

import (
	"context"
	"log/slog"
	"os"
	"time"
)

const (
	LevelDebug slog.Level = -4
	LevelInfo  slog.Level = 0
	LevelWarn  slog.Level = 4
	LevelError slog.Level = 8
)

type Logger struct {
	inner *slog.Logger
}

func NewLogger(h slog.Handler) *Logger {
	return &Logger{
		inner: slog.New(h),
	}
}

func (l *Logger) Log(level slog.Level, msg string, attrs ...any) {
	l.inner.Log(context.Background(), level, msg, attrs...)
}

func (l *Logger) Handler() slog.Handler {
	return l.inner.Handler()
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		l.inner.With(args...),
	}
}

func (l *Logger) New(args ...any) *Logger {
	return l.With(args...)
}

func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.inner.Enabled(ctx, level)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Log(context.Background(), LevelDebug, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.inner.Log(context.Background(), LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Log(context.Background(), LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.inner.Log(context.Background(), LevelError, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.inner.Log(context.Background(), LevelError, msg, args...)
	time.Sleep(time.Second * 3)
	os.Exit(1)
}
