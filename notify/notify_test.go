package notify

import (
	"fmt"
	"testing"
)

func TestEmail(t *testing.T) {
	email := Email{
		Host:      "smtp.qq.com",
		Port:      465,
		Username:  "shuiziliu2022@foxmail.com",
		Password:  "euypgokzavkseceh",
		Recipient: []string{"21723614@qq.com"},
	}
	email.Send("title", "content")
}

func TestWxPusher(t *testing.T) {
	email := WxPusher{
		AppToken: "AT_dWk1PSaCmPieZ8MkuY7KqsOHxuwARB3t",
		TopicIds: []int{31179},
	}
	fmt.Println(email)
	email.Send("title", "content")
}

/*
7977501962:AAGqU5uEsTJZqvNWTWt-9EBrmlRv5CwpvQM
*/
