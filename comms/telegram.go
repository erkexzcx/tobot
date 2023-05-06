package comms

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

var telegramBot *tb.Bot

var replacer strings.Replacer = *strings.NewReplacer(
	"_", "\\_",
	"*", "\\*",
	"[", "\\[",
	"]", "\\]",
	"(", "\\(",
	")", "\\)",
	"~", "\\~",
	"`", "\\`",
	">", "\\>",
	"#", "\\#",
	"+", "\\+",
	"-", "\\-",
	"=", "\\=",
	"|", "\\|",
)

func SendMessageToTelegram(message string) {
	sendTelegramMessage(message)
}

func ForwardMessageToTelegram(rawMessage string, nick string, received bool) {
	telegramMessage := formatForwardableTelegramMessage(rawMessage, nick, received)
	sendTelegramMessage(telegramMessage)
}

func formatForwardableTelegramMessage(rawMessage string, nick string, received bool) string {
	sanitizedMessage := replacer.Replace(rawMessage)
	if received {
		return fmt.Sprintf("*Received from %s:*\n_%s_", nick, sanitizedMessage)
	} else {
		return fmt.Sprintf("*Sent to %s:*\n_%s_", nick, sanitizedMessage)
	}
}

func sendTelegramMessage(msg string) {
	telegramBot.Send(
		&tb.Chat{ID: appConfig.Telegram.ChatId},
		msg,
		&tb.SendOptions{
			ParseMode: tb.ModeMarkdownV2,
		},
	)
}
