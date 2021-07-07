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
