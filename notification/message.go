package notification

type Message interface {
	Subject() string
	HTML() string
	Markdown() string
}
