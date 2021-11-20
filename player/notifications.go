package player

import (
	"fmt"
	"log"
	"strings"
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
