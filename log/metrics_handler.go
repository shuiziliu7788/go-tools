package log

import (
	"context"
	"fmt"
	"github.com/shuiziliu7788/go-tools/notification"
	"log/slog"
	"sync"
	"time"
)

type MetricsHandlerOptions struct {
	Level          slog.Level
	Evaluate       time.Duration
	For            time.Duration
	Expr           string // eq gt egt lt egt 默认 egt
	Threshold      int64
	RepeatInterval time.Duration // 通知的重复间隔
	Notification   func(alert notification.Message) error
}

type MetricsHandler struct {
	slog.Handler
	opts             *MetricsHandlerOptions
	uncounted        int64
	counted          int64
	nextEvaluation   time.Time  // 下次评价时间
	lastEvaluation   time.Time  // 异常开始时间
	lastNotification time.Time  // 上次通知时间
	mu               sync.Mutex // 锁
}

func (m *MetricsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fmt.Println("创建信息")
	return &MetricsHandler{
		Handler: m.Handler.WithAttrs(attrs),
		opts:    m.opts,
	}
}

func (m *MetricsHandler) WithGroup(name string) slog.Handler {
	return &MetricsHandler{
		Handler: m.Handler.WithGroup(name),
		opts:    m.opts,
	}
}

func (m *MetricsHandler) Handle(ctx context.Context, record slog.Record) error {
	if err := m.Handler.Handle(ctx, record); err != nil {
		return err
	}
	m.mu.Lock()
	m.uncounted++
	if m.nextEvaluation.After(record.Time) {
		m.mu.Unlock()
		return nil
	}
	m.counted = m.uncounted
	m.uncounted = 0
	m.nextEvaluation = record.Time.Add(m.opts.Evaluate)
	m.mu.Unlock()
	var firing bool
	switch m.opts.Expr {
	case "eq":
		firing = m.counted == m.opts.Threshold
	case "gt":
		firing = m.counted > m.opts.Threshold
	case "lt":
		firing = m.counted < m.opts.Threshold
	case "egt":
		firing = m.counted <= m.opts.Threshold
	default:
		firing = m.counted >= m.opts.Threshold
	}

	// 无警报
	switch firing {
	case false:
		if m.lastEvaluation.IsZero() {
			return nil
		}
		// 判断是否需要发送恢复通知
		if !m.lastNotification.IsZero() && record.Time.Sub(m.lastEvaluation) > m.opts.For {
			// 发送恢复通知
			go m.opts.Notification(&Alert{
				Status:   false,
				Label:    "",
				Value:    0,
				Record:   slog.Record{},
				StartsAt: time.Time{},
				EndsAt:   time.Time{},
			})
		}
		m.lastEvaluation = time.Time{}
		m.lastNotification = time.Time{}
	default:
		if m.lastEvaluation.IsZero() {
			m.lastEvaluation = record.Time
			return nil
		}
		// 判断持续时间
		if record.Time.Sub(m.lastEvaluation) < m.opts.For {
			return nil
		}
		// 判断是否需要发送通知
		if record.Time.Sub(m.lastNotification) < m.opts.RepeatInterval {
			return nil
		}
		// 发送异常通知
		go m.opts.Notification(Alert{
			Status:   false,
			Label:    "",
			Value:    0,
			Record:   slog.Record{},
			StartsAt: time.Time{},
			EndsAt:   time.Time{},
		})
		// 记录发送通知时间
		m.lastNotification = record.Time
	}

	return nil
}

func NewMetricsHandler(h slog.Handler, opts *MetricsHandlerOptions) *MetricsHandler {
	return &MetricsHandler{
		Handler:        h,
		opts:           opts,
		nextEvaluation: time.Now().Add(opts.Evaluate),
	}
}

type Alert struct {
	Status   bool
	Label    string
	Value    int64
	Record   slog.Record
	StartsAt time.Time
	EndsAt   time.Time
}

func (a Alert) Subject() string {
	return ""
}

func (a Alert) HTML() string {
	return ""
}

func (a Alert) Markdown() string {
	return ""
}
