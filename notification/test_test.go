package notification

import "testing"

type TestMSG struct {
}

func (t TestMSG) Subject() string {
	return "subject"
}

func (t TestMSG) HTML() string {
	return "Html"
}

func (t TestMSG) Markdown() string {
	return "markdown"
}

func TestWxPusher(t *testing.T) {
	pusher := WxPusher{
		AppToken: "AT_1ISyFZ7hYwMtnmO1wJZY9HMQBDkEuQ23",
		TopicIds: []int{31179},
	}
	err := pusher.Send(&TestMSG{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmail(t *testing.T) {
	pusher := Email{
		Host:      "smtp.qq.com",
		Port:      465,
		Username:  "alert_email@qq.com",
		Password:  "rvundgjxiblocabg",
		Recipient: []string{"21723614@qq.com"},
		dialer:    nil,
	}
	err := pusher.Send(&TestMSG{})
	if err != nil {
		t.Fatal(err)
	}
}
