package log

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	defaultNotifications = &notifications{}
	HTML, _              = template.New("example").Parse(`
<div style="background-color:#111217;margin: 0;padding: 0">
    <div style="background:#22252b;background-color:#22252b;margin:0 auto;max-width:600px;min-height: 200px;padding: 20px">
        <div style="text-align: left;border-bottom:1px solid #2f3037;direction:ltr;font-size:0;padding:10px 0;">
            <strong style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;text-align:left;color:#FFFFFF;line-height: 32px;word-break:break-word;">
                {{.Subject}}
            </strong>
        </div>
        <div>
            <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;text-align:left;color:#FFFFFF;line-height: 32px">
                <strong>标签</strong>
            </div>
            <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:12px;line-height:1.5;text-align:left;color:#FFFFFF;word-break:break-word;">
                <p>报警名称: {{.Job}}</p>
                <p>开始时间: {{.StartsAt.Format "2006-01-02 15:04:05"}}</p>
                <p>持续时长: {{.For}}</p>
				<p>错误总计: {{.Value}}</p>
                {{if .Status}}
                <p>恢复时间: {{.EndsAt.Format "2006-01-02 15:04:05"}}</p>
               	{{end}}
            </div>
        </div>
        <div>
            <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;text-align:left;color:#FFFFFF;line-height: 42px">
                <strong>信息</strong>
            </div>
            <div style="background-color:#111217;border:1px solid #2f3037;vertical-align:top;padding:16px;">
                <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;line-height:1.5;text-align:left;color:#FFFFFF;word-break:break-word;">
                    {{if .Status}}
                    当前已恢复
                    {{else}}
                    <p>{{.Record.Message}}</p>
                    {{range .Attrs}}
                    <p>{{.}}</p>
                    {{end}}
                    {{end}}
                </div>
            </div>
        </div>

        <div style="text-align: center;border-top:1px solid #2f3037;direction:ltr;font-size:0;padding:10px 0;">
            <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;line-height:1.5;text-align:left;color:#91929e;">
                故障已持续{{.For}} 开始于 {{.StartsAt.Format "2006-01-02 15:04:05"}}
            </div>
        </div>
    </div>
</div>
`)
)

type MetricsHandlerOptions struct {
	JobName        string
	Level          slog.Level
	Evaluate       time.Duration
	For            time.Duration
	Expr           string // eq gt egt lt egt 默认 egt
	Threshold      int64
	RepeatInterval time.Duration // 通知的重复间隔
	Notifications  Notifier
}

type MetricsHandler struct {
	handler          slog.Handler
	JobName          string
	opts             *MetricsHandlerOptions
	uncounted        int64
	counted          int64
	total            int64
	nextEvaluation   time.Time // 下次评价时间
	lastEvaluation   time.Time // 异常开始时间
	lastNotification time.Time // 上次通知时间
	records          chan slog.Record
	mu               sync.Mutex // 锁
}

func (m *MetricsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.handler.Enabled(ctx, level)
}

func (m *MetricsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MetricsHandler{
		handler:        m.handler.WithAttrs(attrs),
		opts:           m.opts,
		nextEvaluation: time.Now().Add(m.opts.Evaluate),
	}
}

func (m *MetricsHandler) WithGroup(name string) slog.Handler {
	return &MetricsHandler{
		JobName:        fmt.Sprintf("%s/%s", m.JobName, name),
		handler:        m.handler,
		opts:           m.opts,
		nextEvaluation: time.Now().Add(m.opts.Evaluate),
	}
}

func (m *MetricsHandler) Handle(ctx context.Context, record slog.Record) error {
	if err := m.handler.Handle(ctx, record); err != nil {
		return err
	}
	m.mu.Lock()
	if record.Level >= m.opts.Level {
		m.uncounted++
	}
	if m.nextEvaluation.After(record.Time) {
		m.mu.Unlock()
		return nil
	}
	m.nextEvaluation = record.Time.Add(m.opts.Evaluate)
	m.counted = m.uncounted
	m.uncounted = 0
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
	fmt.Println(firing)
	// 无警报
	switch firing {
	case false:
		if m.lastEvaluation.IsZero() {
			return nil
		}
		// 判断是否需要发送恢复通知
		if !m.lastNotification.IsZero() && record.Time.Sub(m.lastEvaluation) > m.opts.For {
			// 发送恢复通知
			go m.opts.Notifications.Send(Alert{
				Status: true,
				Job:    m.JobName,
				Value:  m.total,
				Record: slog.Record{
					Message: "故障已恢复",
				},
				StartsAt: m.lastEvaluation,
				EndsAt:   time.Now(),
			})
		}
		m.lastEvaluation = time.Time{}
		m.lastNotification = time.Time{}
		m.total = 0
	default:
		m.total = m.total + m.counted
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
		go m.opts.Notifications.Send(Alert{
			Status:   false,
			Job:      m.JobName,
			Value:    m.total,
			Record:   record,
			StartsAt: m.lastEvaluation,
			EndsAt:   time.Now(),
		})
		// 记录发送通知时间
		m.lastNotification = record.Time
	}

	return nil
}

func NewMetricsHandler(h slog.Handler, opts *MetricsHandlerOptions) *MetricsHandler {
	if mh, ok := h.(*MetricsHandler); ok {
		h = mh.handler
	}

	if opts.Notifications == nil {
		opts.Notifications = defaultNotifications
	}

	if opts.JobName == "" {
		if execPath, err := os.Executable(); err == nil {
			opts.JobName = filepath.Base(execPath)
		}
	}

	return &MetricsHandler{
		JobName:        opts.JobName,
		handler:        h,
		opts:           opts,
		nextEvaluation: time.Now().Add(opts.Evaluate),
	}
}

type Alert struct {
	Status   bool
	Job      string
	Value    int64
	Record   slog.Record
	StartsAt time.Time
	EndsAt   time.Time
	Attrs    []slog.Attr
}

func (a Alert) Subject() string {
	var builder strings.Builder
	if a.Status {
		builder.WriteString("【已恢复】")
	} else {
		builder.WriteString("【报警】")
	}
	builder.WriteString(a.Job)
	builder.WriteString("(" + a.Record.Message + ")")
	return builder.String()
}

func (a Alert) formatDate(t time.Time, format string) string {
	return t.Format(format)
}

func (a Alert) HTML() string {
	builder := &strings.Builder{}
	a.Record.Attrs(func(attr slog.Attr) bool {
		a.Attrs = append(a.Attrs, attr)
		return true
	})
	if err := HTML.Execute(builder, a); err != nil {
		fmt.Println("错误", err)
		return ""
	}
	return builder.String()
}

func (a Alert) For() time.Duration {
	return a.EndsAt.Sub(a.StartsAt)
}

func (a Alert) Markdown() string {
	return ""
}

type notifications struct {
}

func (n *notifications) Send(alert Alert) error {
	return nil
}
