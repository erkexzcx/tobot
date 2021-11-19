package player

import (
	"sync"
	"time"
	"tobot"
)

type PlayerSettings struct {
	Nick string
	Pass string

	ToTelegram   chan string
	FromTelegram chan string

	MinRTT time.Duration

	RootLink        string
	HeaderUserAgent string
	HeaderHost      string

	BecomeOffline          bool
	BecomeOfflineEveryFrom time.Duration
	BecomeOfflineEveryTo   time.Duration
	BecomeOfflineForFrom   time.Duration
	BecomeOfflineForTo     time.Duration

	RandomizeWait     bool
	RandomizeWaitFrom time.Duration
	RandomizeWaitTo   time.Duration

	Activities []*tobot.Activity
}

func NewPlayer(ps *PlayerSettings) *Player {
	p := &Player{
		nick: ps.Nick,
		pass: ps.Pass,

		toTelegram:   ps.ToTelegram,
		fromTelegram: ps.FromTelegram,

		minRTT: ps.MinRTT,

		rootLink:        ps.RootLink,
		headerUserAgent: ps.HeaderUserAgent,
		headerHost:      ps.HeaderHost,

		becomeOfflineEveryFrom: ps.BecomeOfflineEveryFrom,
		becomeOfflineEveryTo:   ps.BecomeOfflineEveryTo,

		becomeOfflineForFrom: ps.BecomeOfflineForFrom,
		becomeOfflineForTo:   ps.BecomeOfflineForTo,

		randomizeWaitFrom: ps.RandomizeWaitFrom,
		randomizeWaitTo:   ps.RandomizeWaitTo,
	}

	p.timeUntilMux = sync.Mutex{}

	// Update offline times
	p.manageBecomeOffline()

	p.NotifyTelegramSilent("Started!")

	return p
}

type Player struct {
	nick string
	pass string

	toTelegram   chan string
	fromTelegram chan string

	minRTT time.Duration

	rootLink        string
	headerUserAgent string
	headerHost      string

	becomeOfflineEveryFrom time.Duration
	becomeOfflineEveryTo   time.Duration

	becomeOfflineForFrom time.Duration
	becomeOfflineForTo   time.Duration

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
