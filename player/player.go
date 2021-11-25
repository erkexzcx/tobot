package player

import (
	"sync"
	"time"
)

// Set values to these variables before using this package
var (
	MIN_RTT time.Duration
)

type Player struct {
	nick string
	pass string

	rootAddress     string
	headerHost      string
	headerUserAgent string

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
	replyScheduled  map[string]string
	replyMux        sync.Mutex
	waitingForReply bool // Player should freeze until reply is received
}

func NewPlayer(
	nick string,
	pass string,
	rootAddress string,
	headerHost string,
	headerUserAgent string,
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

	// Init reply mechanism (for incoming replies via Telegram)
	go p.listenTelegramMessages(fromTelegram)

	return p
}
