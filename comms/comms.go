package comms

import (
	"log"
	"time"
	"tobot/config"

	"github.com/sashabaranov/go-openai"
	tb "gopkg.in/tucnak/telebot.v2"
)

var appConfig *config.Config

func InitComms(c *config.Config) {
	appConfig = c

	// Connect to Telegram bot
	var err error
	telegramBot, err = tb.NewBot(tb.Settings{
		Token:  c.Telegram.ApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln("Failed to connect to Telegram bot:", err)
	}

	// Create OpenAI client
	openaiClient = openai.NewClient(c.OpenAI.ApiKey)
}
