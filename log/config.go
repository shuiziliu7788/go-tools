package log

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	durationStr := time.Duration(d).String()
	return json.Marshal(durationStr)
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var durationStr string
	if err := json.Unmarshal(data, &durationStr); err != nil {
		return err
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

type Metrics struct {
	JobName        string     `json:"job_name,omitempty"`
	Level          slog.Level `json:"level,omitempty"`
	Evaluate       Duration   `json:"evaluate,omitempty"`
	For            Duration   `json:"for,omitempty"`
	Expr           string     `json:"expr,omitempty"` // eq gt egt lt egt 默认 egt
	Threshold      int64      `json:"threshold,omitempty"`
	RepeatInterval Duration   `json:"repeat_interval,omitempty"` // 通知的重复间隔
	Notify         *Notify    `json:"notify,omitempty"`
}

type Config struct {
	Output      string     `json:"output,omitempty"`
	Level       slog.Level `json:"level,omitempty"`
	Limit       int64      `json:"limit,omitempty"`
	MaxBackups  int        `json:"max_backups,omitempty"`
	RotateHours uint       `json:"rotate_hours,omitempty"`
	Metrics     *Metrics   `json:"metrics,omitempty"`
}

func (c *Config) Logger(setDefault bool) *Logger {
	var handler slog.Handler
	if c.Output != "" {
		if c.Limit == 0 {
			c.Limit = 200
		}
		if c.MaxBackups == 0 {
			c.MaxBackups = 20
		}
		if c.RotateHours == 0 {
			c.RotateHours = 24
		}
		handler = NewTextHandler(NewAsyncFileWriter(&FileHandlerOptions{
			FilePath:    c.Output,
			Limit:       c.Limit,
			MaxBackups:  c.MaxBackups,
			RotateHours: c.RotateHours,
		}), &slog.HandlerOptions{
			Level: c.Level,
		})
	} else {
		handler = NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: c.Level,
		})
	}
	if c.Metrics != nil && c.Metrics.Notify != nil {
		handler = NewMetricsHandler(handler, &MetricsHandlerOptions{
			JobName:        c.Metrics.JobName,
			Level:          c.Metrics.Level,
			Evaluate:       time.Duration(c.Metrics.Evaluate),
			For:            time.Duration(c.Metrics.For),
			Expr:           c.Metrics.Expr,
			Threshold:      c.Metrics.Threshold,
			RepeatInterval: time.Duration(c.Metrics.RepeatInterval),
			Notifications:  c.Metrics.Notify,
		})
	}
	logger := NewLogger(handler)
	if setDefault {
		SetDefault(logger)
	}
	return logger
}
