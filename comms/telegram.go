package comms

import (
	"fmt"
	"log"
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
	"{", "\\{",
	"}", "\\}",
	".", "\\.",
	"!", "\\!",
)

func SendMessageToTelegram(rawMessage string) {
	sanitizedMessage := replacer.Replace(rawMessage)
	sendTelegramMessage(sanitizedMessage)
}

func ForwardMessageToTelegram(rawMessage string, rawNick string, messageReceived bool) {
	telegramMessage := formatForwardableTelegramMessage(rawMessage, rawNick, messageReceived)
	sendTelegramMessage(telegramMessage)
}

func formatForwardableTelegramMessage(rawMessage string, rawNick string, messageReceived bool) string {
	sanitizedMessage := replacer.Replace(rawMessage)
	sanitizedNick := replacer.Replace(rawNick)
	if messageReceived {
		return fmt.Sprintf("*Received from %s:*\n%s", sanitizedNick, sanitizedMessage)
	} else {
		return fmt.Sprintf("*Sent to %s:*\n_%s_", sanitizedNick, sanitizedMessage)
	}
}

func sendTelegramMessage(msg string) {
	_, err := telegramBot.Send(
		&tb.Chat{ID: appConfig.Telegram.ChatId},
		msg,
		&tb.SendOptions{
			ParseMode: tb.ModeMarkdownV2,
		},
	)
	if err != nil {
		log.Println("Failed to send message to Telegram:", err.Error())
	}
}
