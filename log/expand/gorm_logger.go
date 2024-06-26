package expand

import (
	"context"
	"github.com/shuiziliu7788/go-tools/log"
	"time"
)

type GormLogger interface {
	LogMode(level glog.LogLevel) GormLogger
	Info(context.Context, string, ...interface{})
	Warn(context.Context, string, ...interface{})
	Error(context.Context, string, ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
}

type gormLogger struct {
	inner *log.Logger
}

func (g *gormLogger) LogMode(i glog.LogLevel) GormLogger {
	level := log.LevelInfo
	// 处理信息
	switch i {
	case 1, 2:
		level = log.LevelError
	case 3:
		level = log.LevelWarn
	default:
		level = log.LevelInfo
	}

	return &gormLogger{
		inner: log.NewLogger(log.NewLevelHandler(level, g.inner.Handler())),
	}
}

func (g *gormLogger) Info(_ context.Context, msg string, args ...any) {
	g.inner.Info(msg, args)
}

func (g *gormLogger) Warn(_ context.Context, msg string, args ...any) {
	g.inner.Warn(msg, args)
}

func (g *gormLogger) Error(_ context.Context, msg string, args ...any) {
	g.inner.Error(msg, args)
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	//elapsed := time.Since(begin)

	//switch {
	//case err != nil && g.inner.Enabled(ctx, LevelError) && err.Error() != "record not found":
	//	sql, rows := fc()
	//	if rows == -1 {
	//		g.inner.Error("", "sql", sql, "elapsed", elapsed, "err", err)
	//	} else {
	//		g.inner.Error("", "sql", sql, "elapsed", elapsed, "rows", rows, "err", err)
	//	}
	//case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
	//	sql, rows := fc()
	//	slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
	//	if rows == -1 {
	//		l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
	//	} else {
	//		l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	//	}
	//case l.LogLevel == Info:
	//	sql, rows := fc()
	//	if rows == -1 {
	//		l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
	//	} else {
	//		l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	//	}
	//}
}

func NewGormLogger(log *log.Logger) GormLogger {
	return &gormLogger{
		inner: log,
	}
}
