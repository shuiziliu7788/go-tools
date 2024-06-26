package log

import (
	"log/slog"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	handler := NewAsyncFileWrite(&FileHandlerOptions{
		FilePath:    "",
		Limit:       200,
		MaxBackups:  1,
		RotateHours: 6,
	})
	defer handler.Stop()
	l := NewLogger(NewMetricsHandler(slog.NewTextHandler(handler, nil), &MetricsHandlerOptions{
		Level:          slog.LevelError,
		Evaluate:       time.Second,
		For:            time.Minute,
		Expr:           "eq",
		Threshold:      10,
		RepeatInterval: time.Minute,
		Notification: func(alert Alert) {
		},
	}))
	l.Error("SSS")
	l.Error("SSS")
}
