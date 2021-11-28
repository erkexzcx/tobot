package module

import (
	"tobot/player"
)

type Module interface {
	Perform(p *player.Player, settings map[string]string) *Result
	Validate(settings map[string]string) error
}

var Modules = map[string]Module{}

func Add(name string, m Module) {
	Modules[name] = m
}

type Result struct {
	CanRepeat bool  // 'true' if OK, 'false' if inventory is full or resources (needed for activity) has depleted
	Error     error // E.g. banned or anything else unexpected
}
