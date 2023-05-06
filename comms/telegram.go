package comms

import (
	"fmt"
	"log"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

var telegramBot *tb.Bot

var replacer strings.Replacer = *strings.NewReplacer(
	".", "\\.",
	"!", "\\!",
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

func ForwardMessageToTelegram(rawMessage string, rawNick string, messageReceived bool) {
	telegramMessage := formatForwardableTelegramMessage(rawMessage, rawNick, messageReceived)
	err := sendTelegramMessage(telegramMessage)
	if err != nil {
		log.Println("Failed to send message to Telegram:", err.Error())
		sendTelegramMessage("Failed to send message to Telegram: " + err.Error())
	}
}

func formatForwardableTelegramMessage(rawMessage string, rawNick string, messageReceived bool) string {
	sanitizedMessage := replacer.Replace(rawMessage)
	sanitizedNick := replacer.Replace(rawNick)
	if messageReceived {
		return fmt.Sprintf("*Received from %s:*\n_%s_", sanitizedNick, sanitizedMessage)
	} else {
		return fmt.Sprintf("*Sent to %s:*\n%s", sanitizedNick, sanitizedMessage)
	}
}

func sendTelegramMessage(msg string) error {
	_, err := telegramBot.Send(
		&tb.Chat{ID: appConfig.Telegram.ChatId},
		msg,
		&tb.SendOptions{
			ParseMode: tb.ModeMarkdownV2,
		},
	)
	return err
}
