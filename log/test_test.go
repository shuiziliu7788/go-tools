package log

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	var str = `[
  {
    "type": "wx_pusher",
    "app_token": "",
    "topic_ids": [31179]
  }
]
`
	var n Notify
	if err := json.Unmarshal([]byte(str), &n); err != nil {
		t.Fatal(err)
	}
	l := NewLogger(NewMetricsHandler(NewTextHandler(os.Stdout, nil), &MetricsHandlerOptions{
		Level:          LevelError,
		JobName:        "测试系统",
		Evaluate:       3 * time.Second,
		For:            1 * time.Second,
		Threshold:      2,
		RepeatInterval: 5 * time.Second,
		Notifications:  &n,
	}))
	go func() {
		for i := 0; i < 10; i++ {
			l.Error("获取区块错误", "err", "就是获取不到")
			time.Sleep(time.Second)
		}
		for i := 0; i < 10; i++ {
			l.Info("获取区块错误")
			time.Sleep(time.Second)
		}
	}()
	go func() {
		log := l.WithGroup("二级")
		for i := 0; i < 10; i++ {
			log.Error("获取区块错误", "err", "二级")
			time.Sleep(time.Second)
		}
		for i := 0; i < 10; i++ {
			log.Info("获取区块错误")
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Hour)
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

func TestConfig(t *testing.T) {
	options := Metrics{
		Evaluate: Duration(time.Second),
	}
	marshal, _ := json.Marshal(options)
	fmt.Println(options)
	var s Metrics
	if err := json.Unmarshal(marshal, &s); err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}
