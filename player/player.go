package player

import (
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type PlayerSettings struct {
	Nick string
	Pass string

	MinRTTTime time.Duration

	TelegramBot  *tb.Bot
	TelegramChat *tb.Chat

	RootLink        string // Defaults to "http://tob.lt"
	HeaderUserAgent string // Defaults to "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	HeaderHost      string // Defaults to "tob.lt"

	BecomeOffline bool

	BecomeOfflineEveryFrom time.Duration // Defaults to 1h
	BecomeOfflineEveryTo   time.Duration // Defaults to 3h

	BecomeOfflineForFrom time.Duration // Defaults to 15m
	BecomeOfflineForTo   time.Duration // Defaults to 30m
}

func NewPlayer(ps *PlayerSettings) *Player {
	p := &Player{
		nick: ps.Nick,
		pass: ps.Pass,

		minRTTTime: ps.MinRTTTime,

		telegramBot:  ps.TelegramBot,
		telegramChat: ps.TelegramChat,

		rootLink:        ps.RootLink,
		headerUserAgent: ps.HeaderUserAgent,
		headerHost:      ps.HeaderHost,

		becomeOffline: ps.BecomeOffline,

		becomeOfflineEveryFrom: ps.BecomeOfflineEveryFrom,
		becomeOfflineEveryTo:   ps.BecomeOfflineEveryTo,

		becomeOfflineForFrom: ps.BecomeOfflineForFrom,
		becomeOfflineForTo:   ps.BecomeOfflineForTo,
	}

	if ps.HeaderUserAgent == "" {
		p.headerUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	}
	if ps.HeaderHost == "" {
		p.headerHost = "tob.lt"
	}
	if ps.BecomeOfflineEveryFrom == 0 {
		p.becomeOfflineEveryFrom = time.Duration(1 * time.Hour)
	}
	if ps.BecomeOfflineEveryTo == 0 {
		p.becomeOfflineEveryTo = time.Duration(3 * time.Hour)
	}
	if ps.BecomeOfflineForFrom == 0 {
		p.becomeOfflineForFrom = time.Duration(15 * time.Minute)
	}
	if ps.BecomeOfflineForTo == 0 {
		p.becomeOfflineForTo = time.Duration(30 * time.Minute)
	}

	p.waitingPM = false
	p.waitingPMMux = sync.Mutex{}

	p.timeUntilMux = sync.Mutex{}

	// Update offline times
	p.manageBecomeOffline()

	p.initTelegram()

	p.NotifyTelegramSilent("Started!")

	return p
}

type Player struct {
	nick string
	pass string

	minRTTTime time.Duration

	telegramBot  *tb.Bot
	telegramChat *tb.Chat

	rootLink        string
	headerUserAgent string
	headerHost      string

	becomeOffline bool

	becomeOfflineEveryFrom time.Duration
	becomeOfflineEveryTo   time.Duration

	becomeOfflineForFrom time.Duration
	becomeOfflineForTo   time.Duration

	// Used for tracking click times (prevent clicking too fast)
	timeUntilNavigation time.Time
	timeUntilAction     time.Time
	timeUntilMux        sync.Mutex

	// Used to automatically sleep & wakeup
	sleepFrom time.Time
	sleepTo   time.Time

	// For replying to PM
	waitingPMMux sync.Mutex
	waitingPM    bool
}
