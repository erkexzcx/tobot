package player

import (
	"log"
	"strings"
	"tobot/telegram"
)

func (p *Player) Println(v ...interface{}) {
	log.Println("'"+p.nick+"' says:", v)
}

func (p *Player) NotifyTelegram(msg string, silent bool) {
	telegram.SendMessage(telegram.FormatMessage(p.nick, msg), silent)
}

// Format '<send_to>|<message>'
func (p *Player) listenTelegramMessages(ch chan string) {
	for {
		str := <-ch
		strParts := strings.SplitN(str, "|", 2)
		p.replyMux.Lock()
		p.replyScheduled[strParts[0]] = strParts[1]
		p.waitingForReply = false
		p.replyMux.Unlock()
	}
}
