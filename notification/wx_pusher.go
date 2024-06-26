package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WxPusher struct {
	AppToken string   `json:"app_token,omitempty"`
	TopicIds []int    `json:"topic_ids,omitempty"`
	Uids     []string `json:"uids,omitempty"`
}

type request struct {
	AppToken      string   `json:"appToken"`
	Content       string   `json:"content"`
	Summary       string   `json:"summary"`
	ContentType   int      `json:"contentType"`
	TopicIds      []int    `json:"topicIds"`
	Uids          []string `json:"uids"`
	Url           string   `json:"url"`
	VerifyPay     bool     `json:"verifyPay"`
	VerifyPayType int      `json:"verifyPayType"`
}

type result struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
}

func (wx *WxPusher) Send(msg Message) error {
	m := request{
		AppToken:      wx.AppToken,
		Content:       msg.HTML(),
		Summary:       msg.Subject(),
		ContentType:   2,
		TopicIds:      wx.TopicIds,
		Uids:          wx.Uids,
		VerifyPay:     false,
		VerifyPayType: 0,
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal mssage error:%v", err)
	}
	req, err := http.NewRequest("POST", "https://wxpusher.zjiecode.com/api/send/message", bytes.NewBuffer(marshal))
	if err != nil {
		return fmt.Errorf("new request error:%v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message error:%v", err)
	}
	return resp.Body.Close()
}
