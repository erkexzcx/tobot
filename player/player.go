package player

import (
	"sync"
	"time"
)

type Player struct {
	nick string
	pass string

	rootAddress     string
	headerHost      string
	headerUserAgent string
	minRTT          time.Duration

	// Settings
	becomeOfflineEveryFrom time.Duration
	becomeOfflineEveryTo   time.Duration
	becomeOfflineForFrom   time.Duration
	becomeOfflineForTo     time.Duration
	randomizeWaitFrom      time.Duration
	randomizeWaitTo        time.Duration

	// Used for tracking click times (prevent clicking too fast)
	timeUntilNavigation time.Time
	timeUntilAction     time.Time

	// Used for becomeOffline
	sleepFrom time.Time
	sleepTo   time.Time

	// Needed for replies
	replyScheduled map[string]string
	replyMux       sync.Mutex
}

func NewPlayer(
	nick string,
	pass string,
	rootAddress string,
	headerHost string,
	headerUserAgent string,
	minRTT time.Duration,
	fromTelegram chan string,
	becomeOfflineEveryFrom time.Duration,
	becomeOfflineEveryTo time.Duration,
	becomeOfflineForFrom time.Duration,
	becomeOfflineForTo time.Duration,
	randomizeWaitFrom time.Duration,
	randomizeWaitTo time.Duration,
) *Player {
	p := &Player{
		nick: nick,
		pass: pass,

		rootAddress:     rootAddress,
		headerHost:      headerHost,
		headerUserAgent: headerUserAgent,
		minRTT:          minRTT,

		replyScheduled: make(map[string]string),
		replyMux:       sync.Mutex{},
	}

	if becomeOfflineEveryTo != 0 && becomeOfflineForTo != 0 {
		p.becomeOfflineEveryFrom = becomeOfflineEveryFrom
		p.becomeOfflineEveryTo = becomeOfflineEveryTo
		p.becomeOfflineForFrom = becomeOfflineForFrom
		p.becomeOfflineForTo = becomeOfflineForTo
	}

	if randomizeWaitTo != 0 {
		p.randomizeWaitFrom = randomizeWaitFrom
		p.randomizeWaitTo = randomizeWaitTo
	}

	// Init becomeOffline from/to fields
	p.manageBecomeOffline()

	// If service is restarted, we get lots of "too fast" messages, let's wait before first click
	p.timeUntilNavigation = time.Now().Add(MIN_WAIT_TIME - p.minRTT)
	p.timeUntilAction = p.timeUntilNavigation

	return p
}
