package player

import (
	"log"
	"regexp"
	"strings"
	"sync"

	tb "gopkg.in/tucnak/telebot.v2"
)

var reParseReply = regexp.MustCompile(`^User '[*]*([a-zA-Z0-9]+)' says: .+$`)

var (
	paused    = false
	pausedMux = &sync.Mutex{}
)

func getPausedState() bool {
	pausedMux.Lock()
	defer pausedMux.Unlock()
	return paused
}

func setPausedState(state bool) {
	pausedMux.Lock()
	defer pausedMux.Unlock()
	paused = state
}

func (p *Player) initTelegram() {
	p.telegramBot.Handle("/start", func(m *tb.Message) {
		if getPausedState() {
			setPausedState(false)
			p.telegramBot.Send(m.Sender, "player set to resumed")
			return
		}
		p.telegramBot.Send(m.Sender, "player already set to resumed")
	})

	p.telegramBot.Handle("/stop", func(m *tb.Message) {
		if !getPausedState() {
			setPausedState(true)
			p.telegramBot.Send(m.Sender, "player set to paused")
			return
		}
		p.telegramBot.Send(m.Sender, "player already set to paused")
	})

	// Handle PM replies
	p.telegramBot.Handle(tb.OnText, func(m *tb.Message) {
		if !m.IsReply() {
			log.Println("Ignoring Telegram message - not a reply")
			return
		}
		if m.ReplyTo.Chat.ID != p.telegramChat.ID {
			log.Println("Ignoring Telegram message - reply from unknown chat")
			return
		}

		match := reParseReply.FindStringSubmatch(m.ReplyTo.Text)

		if len(match) != 2 {
			log.Println("Ignoring Telegram message - replying to unexpected message")
			return
		}

		defer func() {
			p.waitingPMMux.Lock()
			p.waitingPM = false
			p.waitingPMMux.Unlock()
		}()

		if m.Text == "/ignore" {
			p.NotifyTelegram("ignoring... :)")
			return
		}

		if strings.HasPrefix(m.Text, "/") {
			p.NotifyTelegram("unknown, but ignoring... :)")
			return
		}

		p.sendPM(match[1], strings.TrimSpace(m.Text))
		p.NotifyTelegram("replied :)")
	})

	go p.telegramBot.Start()
}

func (p *Player) NotifyTelegram(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{})
}

func (p *Player) NotifyTelegramSilent(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{DisableNotification: true})
}
