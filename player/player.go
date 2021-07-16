package player

import (
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type PlayerSettings struct {
	Nick string
	Pass string

	MinRTT time.Duration

	TelegramBot  *tb.Bot
	TelegramChat *tb.Chat

	RootLink        string
	HeaderUserAgent string
	HeaderHost      string

	BecomeOffline bool

	BecomeOfflineEveryFrom time.Duration
	BecomeOfflineEveryTo   time.Duration

	BecomeOfflineForFrom time.Duration
	BecomeOfflineForTo   time.Duration

	RandomizeWait bool

	RandomizeWaitFrom time.Duration
	RandomizeWaitTo   time.Duration
}

func NewPlayer(ps *PlayerSettings) *Player {
	p := &Player{
		nick: ps.Nick,
		pass: ps.Pass,

		minRTT: ps.MinRTT,

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

		randomizeWait: ps.RandomizeWait,

		randomizeWaitFrom: ps.RandomizeWaitFrom,
		randomizeWaitTo:   ps.RandomizeWaitTo,
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

	minRTT time.Duration

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

	randomizeWait bool

	randomizeWaitFrom time.Duration
	randomizeWaitTo   time.Duration

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
