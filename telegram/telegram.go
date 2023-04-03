package telegram

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Set this variable before using this package
var CHAT_ID int64

var telegramBot *tb.Bot

func Start(out map[string]chan string, tBot *tb.Bot) {
	telegramBot = tBot
	SendMessage("Program started!", true)
	go telegramBot.Start()
}

func FormatMessage(nick string, message string) string {
	return fmt.Sprintf("[%s] %s", nick, message)
}

func SendMessage(msg string, silent bool) {
	telegramBot.Send(&tb.Chat{ID: CHAT_ID}, msg, &tb.SendOptions{DisableNotification: silent})
}
