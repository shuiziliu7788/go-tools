package log

import (
	"os"
	"testing"
)

func TestTrace(t *testing.T) {
	Trace("Trace", "msg", "SSS")
	Debug("Debug", "msg", "SSS")
	Info("Info", "msg", "SSS")
	Warn("Warn", "msg", "SSS")
	Error("Error", "msg", "SSS")
}

func TestMetric(t *testing.T) {
	metric := &Metric{
		lvl: LevelTrace,
	}
	l := NewLogger(NewMetricHandler(NewTerminalHandlerWithLevel(os.Stdout, LevelWarn, true), metric))
	with := l.With("with", "with")

	with.Trace("Trace", "with_msg", "SSS")
	with.Debug("Debug", "with_msg", "SSS")
	with.Info("Info", "with_msg", "SSS")
	with.Warn("Warn", "with_msg", "SSS")
	with.Error("Error", "with_msg", "SSS")
}
