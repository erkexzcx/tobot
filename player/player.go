package player

import (
	"time"
	"tobot/config"
)

type Player struct {
	Config *config.Player

	// Used for tracking clicks (to prevent clicking too fast)
	timeUntilNavigation time.Time
	timeUntilAction     time.Time

	// Used for becomeOffline tracking
	sleepFrom time.Time
	sleepTo   time.Time
}

func NewPlayer(c *config.Player) *Player {
	p := &Player{
		Config: c,
	}

	// Init becomeOffline from/to fields
	p.manageBecomeOffline()

	// Wait before first click (in case software is in restart-loop)
	p.timeUntilNavigation = time.Now().Add(MIN_WAIT_TIME - *p.Config.Settings.MinRTT)
	p.timeUntilAction = p.timeUntilNavigation

	return p
}
