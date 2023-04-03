package player

import (
	"fmt"
	"log"
	"tobot/telegram"
)

func (p *Player) Println(v ...interface{}) {
	str := "[" + p.nick + "]"
	for _, v := range v {
		str += fmt.Sprintf(" %v", v)
	}
	log.Println(str)
}

func (p *Player) NotifyTelegram(msg string, silent bool) {
	telegram.SendMessage(telegram.FormatMessage(p.nick, msg), silent)
}
