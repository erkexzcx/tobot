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

func ForwardMessageToTelegram(myplayer, rawMessage, rawNick string, messageReceived bool) {
	telegramMessage := formatForwardableTelegramMessage(myplayer, rawMessage, rawNick, messageReceived)
	sendTelegramMessage(telegramMessage)
}

func formatForwardableTelegramMessage(myplayer, rawMessage, rawNick string, messageReceived bool) string {
	sanitizedMessage := replacer.Replace(rawMessage)
	sanitizedNick := replacer.Replace(rawNick)
	sanitizedPlayer := replacer.Replace(myplayer)
	if messageReceived {
		return fmt.Sprintf("*Received %s \\-\\> %s:*\n%s", sanitizedNick, sanitizedPlayer, sanitizedMessage)
	} else {
		return fmt.Sprintf("*Sent %s \\-\\> %s:*\n_%s_", sanitizedPlayer, sanitizedNick, sanitizedMessage)
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
