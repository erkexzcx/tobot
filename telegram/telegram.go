package telegram

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// [player1] Player '*moderator' says: hello world!
// [player1] Player 'moderator' says: hello world!
var reParseReply = regexp.MustCompile(`^\[([a-zA-Z0-9]+)\] Player '[*]*([a-zA-Z0-9]+)' says: .+$`)

// Set this variable before using this package
var CHAT_ID int64

var telegramBot *tb.Bot

func Start(out map[string]chan string, tBot *tb.Bot) {
	telegramBot = tBot

	// Handle PM replies
	telegramBot.Handle(tb.OnText, func(m *tb.Message) {
		if !m.IsReply() {
			log.Println("Ignoring Telegram message - not a reply")
			return
		}
		if m.ReplyTo.Chat.ID != CHAT_ID {
			log.Println("Ignoring Telegram message - reply from unknown chat")
			return
		}

		match := reParseReply.FindStringSubmatch(m.ReplyTo.Text)
		if len(match) != 3 {
			log.Println("Ignoring Telegram message - reply to unexpected message")
			return
		}
		if m.Text == "/ignore" {
			SendMessage("ignoring... :)", false)
			return
		}
		if strings.HasPrefix(m.Text, "/") {
			SendMessage("unknown, but ignoring...", false)
			return
		}

		replyFrom := match[1]
		replyTo := match[2]

		ch, found := out[replyFrom]
		if !found {
			SendMessage("unable to find player from which to reply, ignoring...", false)
			return
		}

		ch <- replyTo + "|" + strings.TrimSpace(m.Text)
		SendMessage("reply sent!", false)
	})

	SendMessage("Program started!", false)

	telegramBot.Start()
}

func FormatMessage(nick string, message string) string {
	return fmt.Sprintf("[%s] %s", nick, message)
}

func SendMessage(msg string, silent bool) {
	telegramBot.Send(&tb.Chat{ID: CHAT_ID}, msg, &tb.SendOptions{DisableNotification: silent})
}
