package notification

import "gopkg.in/gomail.v2"

type Email struct {
	Host      string   `json:"host,omitempty"`
	Port      int      `json:"port,omitempty"`
	Username  string   `json:"username,omitempty"`
	Password  string   `json:"password,omitempty"`
	Recipient []string `json:"recipient,omitempty"`
	dialer    *gomail.Dialer
}

func (m *Email) Dialer() {
	m.dialer = gomail.NewDialer(
		m.Host,
		m.Port,
		m.Username,
		m.Password,
	)
}

func (m *Email) Send(msg Message) error {
	m.dialer = gomail.NewDialer(
		m.Host,
		m.Port,
		m.Username,
		m.Password,
	)
	message := gomail.NewMessage()
	message.SetHeader("From", m.dialer.Username)
	message.SetHeader("To", m.Recipient...)
	message.SetHeader("Subject", msg.Subject())
	message.SetBody("text/html", msg.HTML())
	return m.dialer.DialAndSend(message)
}
