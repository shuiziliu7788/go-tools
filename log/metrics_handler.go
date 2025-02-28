package log

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Notify interface {
	Send(title string, content string)
}

type Metric struct {
	Name           string        `json:"name"`
	Level          slog.Level    `json:"level"`
	NotifyPeriod   time.Duration `json:"notify_period"`
	EvaluatePeriod time.Duration `json:"evaluate_period"`
	Threshold      int           `json:"threshold"`
	lastNotify     time.Time
	nextNotifyTime time.Time
	records        []*slog.Record
	mu             sync.Mutex
}

func (m *Metric) sendNotification() {
	m.lastNotify = time.Now()
	m.nextNotifyTime = m.lastNotify.Add(m.NotifyPeriod)

}

func (m *Metric) Handle(record slog.Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.records) < m.Threshold+5 {
		m.records = append(m.records, &record)
		return
	}
	m.records = append(m.records, &record)
	periodAgo := time.Now().Add(-m.EvaluatePeriod)

	for i := len(m.records) - 1; i < 0; i-- {
		if !periodAgo.After(m.records[i].Time) {
			m.records = m.records[:i]
			break
		}
	}
	if len(m.records) < m.Threshold || time.Now().Before(m.nextNotifyTime) {
		return
	}

}

type MetricHandler struct {
	handler slog.Handler
	metric  *Metric
}

func (m *MetricHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.handler.Enabled(ctx, level)
}

func (m *MetricHandler) Handle(ctx context.Context, record slog.Record) error {
	m.metric.Handle(record)
	if m.handler.Enabled(ctx, record.Level) {
		return m.handler.Handle(ctx, record)
	}
	return nil
}

func (m *MetricHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MetricHandler{
		handler: m.handler.WithAttrs(attrs),
		metric:  m.metric,
	}
}

func (m *MetricHandler) WithGroup(name string) slog.Handler {
	return &MetricHandler{
		handler: m.handler.WithGroup(name),
		metric:  m.metric,
	}
}

func NewMetricHandler(h slog.Handler, metric *Metric) *MetricHandler {
	return &MetricHandler{
		handler: h,
		metric:  metric,
	}
}
