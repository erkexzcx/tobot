package main

import (
	"flag"
	"log"
	"net/url"
	"os"
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

	// Parse configuration file
	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Set global variables in package "telegram"
	telegram.CHAT_ID = c.Telegram.ChatId

	// Connect to Telegram bot
	telegramBot, err := tb.NewBot(tb.Settings{
		Token:  c.Telegram.ApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
	}

	parsedRootAddr, err := url.Parse(c.Settings.RootAddress)
	if err != nil {
		log.Fatalln(err)
	}

	var playerBecomeOfflineEnabled, playerRandomizeWaitEnabled bool

	// Parse default values
	var becomeOfflineEveryFrom, becomeOfflineEveryTo, becomeOfflineForFrom, becomeOfflineForTo time.Duration
	val, _ := strconv.ParseBool(c.Settings.BecomeOffline.Enabled)
	if val {
		playerBecomeOfflineEnabled = true
		becomeOfflineEvery := strings.Split(c.Settings.BecomeOffline.Every, ",")
		becomeOfflineFor := strings.Split(c.Settings.BecomeOffline.For, ",")
		becomeOfflineEveryFrom, _ = time.ParseDuration(becomeOfflineEvery[0])
		becomeOfflineEveryTo, _ = time.ParseDuration(becomeOfflineEvery[1])
		becomeOfflineForFrom, _ = time.ParseDuration(becomeOfflineFor[0])
		becomeOfflineForTo, _ = time.ParseDuration(becomeOfflineFor[1])
	}
	var randomizeWaitFrom, randomizeWaitTo time.Duration
	val, _ = strconv.ParseBool(c.Settings.RandomizeWait.Enabled)
	if val {
		playerRandomizeWaitEnabled = true
		randomizeWaitVal := strings.Split(c.Settings.RandomizeWait.WaitVal, ",")
		randomizeWaitFrom, _ = time.ParseDuration(randomizeWaitVal[0])
		randomizeWaitTo, _ = time.ParseDuration(randomizeWaitVal[1])
	}

	// Create map where each player will have it's own channel for messages _to_ players
	outChannels := make(map[string]chan string)

	playersActivities := make(map[*player.Player][]*tobot.Activity)
	for _, playerConfig := range c.Players {
		// "Merge" activity files
		a := make([]*tobot.Activity, 0)
		files, err := filepath.Glob(playerConfig.ActivitiesDir + "/*.yml")
		if err != nil {
			log.Fatalln("Failed to read activities .yml files of player '" + playerConfig.Nick + "': " + err.Error())
		}
		for _, f := range files {
			if strings.HasPrefix(path.Base(f), "_") {
				continue // Skip '_*.yml' files
			}
			contents, err := os.ReadFile(f)
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
		outChannel := make(chan string, 100) // buffered channel
		outChannels[playerConfig.Nick] = outChannel

		// Make everything more simple
		playerNick := playerConfig.Nick
		playerPass := playerConfig.Pass
		playerRootAddress := c.Settings.RootAddress
		playerHeaderHost := parsedRootAddr.Host
		playerUserAgent := c.Settings.UserAgent
		playerMinRTT := c.Settings.MinRTT
		playerFromTelegram := outChannel
		playerBecomeOfflineEveryFrom := becomeOfflineEveryFrom
		playerBecomeOfflineEveryTo := becomeOfflineEveryTo
		playerBecomeOfflineForFrom := becomeOfflineForFrom
		playerBecomeOfflineForTo := becomeOfflineForTo
		playerRandomizeWaitFrom := randomizeWaitFrom
		playerRandomizeWaitTo := randomizeWaitTo
		playerActivities := a

		// Override defalt values for becomeOffline & randomizeWait if specified in player level
		if playerConfig.Settings.RootAddress != "" {
			playerRootAddress = playerConfig.Settings.RootAddress
			tmpParsedRootAddr, err := url.Parse(playerConfig.Settings.RootAddress)
			if err != nil {
				log.Fatalln(err)
			}
			playerHeaderHost = tmpParsedRootAddr.Host
		}
		if playerConfig.Settings.UserAgent != "" {
			playerUserAgent = playerConfig.Settings.UserAgent
		}
		if playerConfig.Settings.MinRTT != 0 {
			playerMinRTT = playerConfig.Settings.MinRTT
		}

		tmpPlayerBecomeOfflineEnabled := playerBecomeOfflineEnabled
		tmpPlayerRandomizeWaitEnabled := playerRandomizeWaitEnabled
		val, err := strconv.ParseBool(playerConfig.Settings.BecomeOffline.Enabled)
		if playerConfig.Settings.BecomeOffline.Enabled != "" && err == nil {
			tmpPlayerBecomeOfflineEnabled = val
		}
		val, err = strconv.ParseBool(playerConfig.Settings.RandomizeWait.Enabled)
		if playerConfig.Settings.RandomizeWait.Enabled != "" && err == nil {
			tmpPlayerRandomizeWaitEnabled = val
		}

		if c.Settings.BecomeOffline.Every != "" {
			becomeOfflineEvery := strings.Split(c.Settings.BecomeOffline.Every, ",")
			playerBecomeOfflineEveryFrom, _ = time.ParseDuration(becomeOfflineEvery[0])
			playerBecomeOfflineEveryTo, _ = time.ParseDuration(becomeOfflineEvery[1])
		}
		if playerConfig.Settings.BecomeOffline.Every != "" {
			becomeOfflineEvery := strings.Split(playerConfig.Settings.BecomeOffline.Every, ",")
			playerBecomeOfflineEveryFrom, _ = time.ParseDuration(becomeOfflineEvery[0])
			playerBecomeOfflineEveryTo, _ = time.ParseDuration(becomeOfflineEvery[1])
		}
		if c.Settings.BecomeOffline.For != "" {
			becomeOfflineFor := strings.Split(c.Settings.BecomeOffline.For, ",")
			playerBecomeOfflineForFrom, _ = time.ParseDuration(becomeOfflineFor[0])
			playerBecomeOfflineForTo, _ = time.ParseDuration(becomeOfflineFor[1])
		}
		if playerConfig.Settings.BecomeOffline.For != "" {
			becomeOfflineFor := strings.Split(playerConfig.Settings.BecomeOffline.For, ",")
			playerBecomeOfflineForFrom, _ = time.ParseDuration(becomeOfflineFor[0])
			playerBecomeOfflineForTo, _ = time.ParseDuration(becomeOfflineFor[1])
		}
		if !tmpPlayerBecomeOfflineEnabled {
			playerBecomeOfflineEveryFrom = 0
			playerBecomeOfflineEveryTo = 0
			playerBecomeOfflineForFrom = 0
			playerBecomeOfflineForTo = 0
		}

		if c.Settings.RandomizeWait.WaitVal != "" {
			randomizeWait := strings.Split(c.Settings.RandomizeWait.WaitVal, ",")
			playerRandomizeWaitFrom, _ = time.ParseDuration(randomizeWait[0])
			playerRandomizeWaitTo, _ = time.ParseDuration(randomizeWait[1])
		}
		if playerConfig.Settings.RandomizeWait.WaitVal != "" {
			randomizeWait := strings.Split(playerConfig.Settings.RandomizeWait.WaitVal, ",")
			playerRandomizeWaitFrom, _ = time.ParseDuration(randomizeWait[0])
			playerRandomizeWaitTo, _ = time.ParseDuration(randomizeWait[1])
		}
		if !tmpPlayerRandomizeWaitEnabled {
			playerRandomizeWaitFrom = 0
			playerRandomizeWaitTo = 0
		}

		p := player.NewPlayer(
			playerNick,
			playerPass,
			playerRootAddress,
			playerHeaderHost,
			playerUserAgent,
			playerMinRTT-time.Millisecond,
			playerFromTelegram,
			playerBecomeOfflineEveryFrom,
			playerBecomeOfflineEveryTo,
			playerBecomeOfflineForFrom,
			playerBecomeOfflineForTo,
			playerRandomizeWaitFrom,
			playerRandomizeWaitTo,
		)

		playersActivities[p] = playerActivities
	}

	telegram.Start(outChannels, telegramBot)

	for p, a := range playersActivities {
		go tobot.Start(p, a)
	}

	select {} // block current routine
}
