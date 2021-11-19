package telegram

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	inChannel      chan string
	outChannels    map[string]chan string
	telegramBot    *tb.Bot
	telegramChatID int64
)

// Takes 'in' as routine which accepts pre-formatted messages from players. E.g. 'player1 says: hello world!'
// Takes 'out' as players map (string=player's nick; chan string=message from the user back to the player)
// Telegram bot is established Telegram bot object
// TelegramChatID is chat ID to which send & accept messages
func Start(in chan string, out map[string]chan string, tBot *tb.Bot, chatID int64) {
	inChannel = in
	outChannels = out
	telegramBot = tBot
	telegramChatID = chatID

	// Listen for messages from users

	telegramBot.Start() // This blocks routine
}

func FormatMessage(fromNick string, message string) string {
	return fmt.Sprintf("'%s' says: %s", fromNick, message)
}

func listenForUserReplies() {

}

func (p *Player) NotifyTelegram(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{})
}

func (p *Player) NotifyTelegramSilent(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{DisableNotification: true})
}
