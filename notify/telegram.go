package notify

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/shuiziliu7788/go-tools/log"
)

type Telegram struct {
	Token  string `json:"token"`
	ChatId int64  `json:"chat_id"`
}

func (t *Telegram) Send(title string, content string) {
	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Error("new bot error", "err", err)
		return
	}
	msg := tgbotapi.NewMessage(t.ChatId, fmt.Sprintf("<b>%s</b> (%s)\n", title, content))
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	if _, err = bot.Send(msg); err != nil {
		log.Error("send telegram message error", "err", err)
	}
}
