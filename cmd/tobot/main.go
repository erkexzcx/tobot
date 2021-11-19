package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tobot"

	"tobot/config"
	"tobot/player"
	"tobot/telegram"

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

	parsedLink, err := url.Parse(c.RootAddress)
	if err != nil {
		log.Fatalln(err)
	}

	telegramBot, err := tb.NewBot(tb.Settings{
		Token:  c.Telegram.ApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
	}

	var becomeOfflineEveryFrom, becomeOfflineEveryTo, becomeOfflineForFrom, becomeOfflineForTo time.Duration
	val, err := strconv.ParseBool(c.Settings.BecomeOffline.Enabled)
	if c.Settings.BecomeOffline.Enabled != "" && err == nil && val {
		becomeOfflineEvery := strings.Split(c.Settings.BecomeOffline.Every, ",")
		becomeOfflineFor := strings.Split(c.Settings.BecomeOffline.For, ",")
		becomeOfflineEveryFrom, _ = time.ParseDuration(becomeOfflineEvery[0])
		becomeOfflineEveryTo, _ = time.ParseDuration(becomeOfflineEvery[1])
		becomeOfflineForFrom, _ = time.ParseDuration(becomeOfflineFor[0])
		becomeOfflineForTo, _ = time.ParseDuration(becomeOfflineFor[1])
	}

	var randomizeWaitFrom, randomizeWaitTo time.Duration
	val, err = strconv.ParseBool(c.Settings.RandomizeWait.Enabled)
	if c.Settings.RandomizeWait.Enabled != "" && err == nil && val {
		randomizeWaitVal := strings.Split(c.Settings.RandomizeWait.WaitVal, ",")
		randomizeWaitFrom, _ = time.ParseDuration(randomizeWaitVal[0])
		randomizeWaitTo, _ = time.ParseDuration(randomizeWaitVal[1])
	}

	// Create channel where telegram goroutine accepts messages _from_ players
	inChannel := make(chan string)

	// Create map where each player will have it's own channel for messages _to_ players
	outChannels := make(map[string]chan string)

	var players []*player.Player
	for _, p := range c.Players {
		// "Merge" activity files
		a := make([]*tobot.Activity, 0)
		files, err := filepath.Glob(*&p.ActivitiesDir + "/*.yml")
		if err != nil {
			log.Fatalln("Failed to read activities .yml files of player '" + p.Nick + "': " + err.Error())
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

		// Create map and store in outChannels
		outChannel := make(chan string)
		outChannels[p.Nick] = outChannel

		ps := &player.PlayerSettings{
			Nick: p.Nick,
			Pass: p.Pass,

			ToTelegram:   inChannel,
			FromTelegram: outChannel,

			MinRTT: c.MinRTT - time.Millisecond,

			RootLink: c.RootAddress,

			HeaderUserAgent: c.Settings.UserAgent,
			HeaderHost:      parsedLink.Host,

			BecomeOfflineEveryFrom: becomeOfflineEveryFrom,
			BecomeOfflineEveryTo:   becomeOfflineEveryTo,

			BecomeOfflineForFrom: becomeOfflineForFrom,
			BecomeOfflineForTo:   becomeOfflineForTo,

			RandomizeWaitFrom: randomizeWaitFrom,
			RandomizeWaitTo:   randomizeWaitTo,

			Activities: a,
		}
		players = append(players, player.NewPlayer(ps))
	}

	for _, p := range players {
		go tobot.Start(p)
	}
	telegram.Start(inChannel, outChannels, telegramBot, c.Telegram.ChatId) // this also blocks main routine
}
