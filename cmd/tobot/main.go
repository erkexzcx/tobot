package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
	"tobot"

	"tobot/config"
	"tobot/player"

	_ "tobot/module/all"

	tb "gopkg.in/tucnak/telebot.v2"
	"gopkg.in/yaml.v2"
)

var configPath = flag.String("config", "config.yml", "path to config file")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	a := make([]*tobot.Activity, 0)
	files, err := filepath.Glob(*activitiesDir + "/*.yml")
	if err != nil {
		panic("Failed to read activities .yml files: " + err.Error())
	}
	for _, f := range files {
		if strings.HasPrefix(path.Base(f), "_") {
			continue // Skip '_*.yml' files
		}
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalln(err)
		}

		var aa *tobot.Activity
		if err := yaml.Unmarshal(contents, &aa); err != nil {
			log.Fatalln(err)
		}
		a = append(a, aa)
	}

	telegramBot, err := tb.NewBot(tb.Settings{
		Token:  c.TelegramApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
		return
	}

	parsedLink, err := url.Parse(c.RootAddress)
	if err != nil {
		log.Fatalln(err)
		return
	}

	becomeOfflineEvery := strings.Split(c.BecomeOfflineEvery, ",")
	becomeOfflineFor := strings.Split(c.BecomeOfflineFor, ",")
	randomizeWaitVal := strings.Split(c.RandomizeWaitVal, ",")

	becomeOfflineEveryFrom, _ := time.ParseDuration(becomeOfflineEvery[0])
	becomeOfflineEveryTo, _ := time.ParseDuration(becomeOfflineEvery[1])
	becomeOfflineForFrom, _ := time.ParseDuration(becomeOfflineFor[0])
	becomeOfflineForTo, _ := time.ParseDuration(becomeOfflineFor[1])

	randomizeWaitFrom, _ := time.ParseDuration(randomizeWaitVal[0])
	randomizeWaitTo, _ := time.ParseDuration(randomizeWaitVal[1])

	ps := &player.PlayerSettings{
		Nick: c.Nick,
		Pass: c.Pass,

		MinRTT: c.MinRTT - time.Millisecond,

		TelegramBot:  telegramBot,
		TelegramChat: &tb.Chat{ID: c.TelegramChatId},

		RootLink: c.RootAddress,

		HeaderUserAgent: c.UserAgent,
		HeaderHost:      parsedLink.Host,

		BecomeOffline: c.BecomeOffline,

		BecomeOfflineEveryFrom: becomeOfflineEveryFrom,
		BecomeOfflineEveryTo:   becomeOfflineEveryTo,

		BecomeOfflineForFrom: becomeOfflineForFrom,
		BecomeOfflineForTo:   becomeOfflineForTo,

		RandomizeWait: c.RandomizeWait,

		RandomizeWaitFrom: randomizeWaitFrom,
		RandomizeWaitTo:   randomizeWaitTo,
	}
	p := player.NewPlayer(ps)

	tobot.Start(p, c, a)
}
