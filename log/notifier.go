package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
	"net/http"
)

type Notifier interface {
	Send(alert Alert) error
}

type Email struct {
	Host      string   `json:"host,omitempty"`
	Port      int      `json:"port,omitempty"`
	Username  string   `json:"username,omitempty"`
	Password  string   `json:"password,omitempty"`
	Recipient []string `json:"recipient,omitempty"`
	dialer    *gomail.Dialer
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var email struct {
		Host      string   `json:"host,omitempty"`
		Port      int      `json:"port,omitempty"`
		Username  string   `json:"username,omitempty"`
		Password  string   `json:"password,omitempty"`
		Recipient []string `json:"recipient,omitempty"`
	}
	if err := json.Unmarshal(data, &email); err != nil {
		return err
	}
	if email.Host == "" {
		return errors.New("email.host is required")
	} else if email.Port == 0 {
		return errors.New("email.port is required")
	} else if email.Username == "" {
		return errors.New("email.username is required")
	} else if email.Password == "" {
		return errors.New("email.password is required")
	} else if email.Recipient == nil || len(email.Recipient) == 0 {
		return errors.New("email.recipient is required")
	}
	e.Host = email.Host
	e.Port = email.Port
	e.Username = email.Username
	e.Password = email.Password
	e.Recipient = email.Recipient
	e.dialer = gomail.NewDialer(
		email.Host,
		email.Port,
		email.Username,
		email.Password,
	)
	return nil
}

func (e *Email) Send(alert Alert) error {
	e.dialer = gomail.NewDialer(
		e.Host,
		e.Port,
		e.Username,
		e.Password,
	)
	message := gomail.NewMessage()
	message.SetHeader("From", e.dialer.Username)
	message.SetHeader("To", e.Recipient...)
	message.SetHeader("Subject", alert.Subject())
	message.SetBody("text/html", alert.HTML())
	return e.dialer.DialAndSend(message)
}

type WxPusher struct {
	AppToken string   `json:"app_token,omitempty"`
	TopicIds []int    `json:"topic_ids,omitempty"`
	Uids     []string `json:"uids,omitempty"`
}

func (wx *WxPusher) UnmarshalJSON(data []byte) error {
	var pusher struct {
		AppToken string   `json:"app_token,omitempty"`
		TopicIds []int    `json:"topic_ids,omitempty"`
		Uids     []string `json:"uids,omitempty"`
	}
	if err := json.Unmarshal(data, &pusher); err != nil {
		return err
	}
	if pusher.AppToken == "" {
		return errors.New("pusher.app_token is required")
	} else if len(pusher.TopicIds) == 0 && len(pusher.Uids) == 0 {
		return errors.New("pusher.topic_ids or pusher.uids is required")
	}
	wx.AppToken = pusher.AppToken
	wx.TopicIds = pusher.TopicIds
	wx.Uids = pusher.Uids
	return nil
}

func (wx *WxPusher) Send(alert Alert) error {
	marshal, err := json.Marshal(map[string]any{
		"appToken":      wx.AppToken,
		"content":       alert.HTML(),
		"summary":       alert.Subject(),
		"contentType":   2,
		"topicIds":      wx.TopicIds,
		"uids":          wx.Uids,
		"verifyPay":     false,
		"verifyPayType": 0,
	})
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

type Notify struct {
	Notifier []Notifier
}

func (n *Notify) Send(alert Alert) error {
	for _, notifier := range n.Notifier {
		notifier.Send(alert)
	}
	return nil
}

func (n *Notify) convert(values any) (Notifier, error) {
	m, ok := values.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid json error:%v", values)
	}
	jsonData, _ := json.Marshal(values)
	switch m["type"] {
	case "email":
		email := &Email{}
		return email, json.Unmarshal(jsonData, email)
	case "wx_pusher":
		wxPusher := &WxPusher{}
		return wxPusher, json.Unmarshal(jsonData, wxPusher)
	default:
		return nil, fmt.Errorf("invalid config type error:%v", m["type"])
	}
}

func (n *Notify) UnmarshalJSON(data []byte) error {
	var values any
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	switch value := values.(type) {
	case map[string]any:
		if notify, err := n.convert(value); err != nil {
			return fmt.Errorf("convert notify error:%v", err)
		} else {
			n.Notifier = append(n.Notifier, notify)
		}
	case []any:
		for _, v := range value {
			if notify, err := n.convert(v); err != nil {
				return fmt.Errorf("convert notify error:%v", err)
			} else {
				n.Notifier = append(n.Notifier, notify)
			}
		}
	default:
		return fmt.Errorf("invalid json type:%v", value)
	}
	return nil
}

func (n *Notify) MarshalJSON() ([]byte, error) {
	if len(n.Notifier) == 0 {
		return nil, nil
	} else if len(n.Notifier) == 1 {
		return json.Marshal(n.Notifier[0])
	}
	return json.Marshal(n.Notifier)
}
