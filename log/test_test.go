package log

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	writer := NewAsyncFileWriter(&FileHandlerOptions{
		FilePath:    "",
		Limit:       200,
		MaxBackups:  1,
		RotateHours: 6,
	})
	defer writer.Stop()
	l := NewLogger(NewMetricsHandler(NewTextHandler(writer, nil), &MetricsHandlerOptions{
		Level:          slog.LevelError,
		Evaluate:       time.Second,
		For:            time.Minute,
		Expr:           "eq",
		Threshold:      10,
		RepeatInterval: time.Minute,
	}))
	l.Error("SSS")
	l.Error("SSS")
}

func TestNotify(t *testing.T) {
	var str = `[
  {
    "type": "wx_pusher",
    "app_token": "AT_s0Q5e8mpbeBvhAdx0WMms0YNxyrWCKV1",
    "topic_ids": [31179]
  },
 {
      "type": "email",
    "host": "smtp.qq.com",
    "port": 465,
    "username": "alert_email@qq.com",
    "password": "rvundgjxiblocabg",
    "recipient": ["21723614@qq.com"],
    "topic_ids": [31179]
  }
]
`

	var n Notify
	if err := json.Unmarshal([]byte(str), &n); err != nil {
		t.Fatal(err)
	}
	alert := Alert{
		Status: false,
		Job:    "ETH",
		Value:  0,
		Record: slog.Record{
			Message: "系统错误",
		},
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}
	alert.StartsAt.Format("2006-01-02 15:04:05 G")
	n.Send(alert)
}

func TestHtml(t *testing.T) {
	alert := Alert{
		Status:   false,
		Job:      "ETH",
		Value:    0,
		Record:   slog.Record{},
		StartsAt: time.Time{},
		EndsAt:   time.Time{},
	}
	fmt.Println(alert.HTML())
}
