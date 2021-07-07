package tobot

import (
	"fmt"
	"log"
	"strconv"
	"tobot/config"
	"tobot/module"
	"tobot/player"
)

type Activity struct {
	Name  string              `yaml:"name"`
	Tasks []map[string]string `yaml:"tasks"`
}

func Start(p *player.Player, c *config.Config, a []*Activity) {
	// Validate activities
	for _, a := range a {
		validateActivity(c, a)
	}

	// Run activities in a loop
	for {
		for _, aa := range a {
			runActivity(p, c, aa)
		}
	}
}

func validateActivity(c *config.Config, a *Activity) {
	for _, task := range a.Tasks {
		count, found := task["_count"]
		if found {
			countInt, err := strconv.Atoi(count)
			if err != nil {
				panic("invalid '_count' value (in '" + a.Name + "' activity)")
			}
			if countInt < 0 {
				panic("invalid '_count' value (in '" + a.Name + "' activity)")
			}
		}
		if task["_module"] == "" {
			panic("task missing _module (in '" + a.Name + "' activity)")
		}
		m, found := module.Modules[task["_module"]]
		if !found {
			panic("unknown _module '" + task["_module"] + "' (in '" + a.Name + "' activity)")
		}
		err := m.Validate(task)
		if err != nil {
			panic("Error from activity '" + a.Name + "' module '" + task["_module"] + "': " + err.Error())
		}
	}
}

func runActivity(p *player.Player, c *config.Config, a *Activity) {
	log.Print("Started '" + a.Name + "'")

	for _, task := range a.Tasks {
		m := module.Modules[task["_module"]]

		count, _ := strconv.Atoi(task["_count"])
		endless := false
		if count == 0 {
			endless = true
		}

		log.Println("module: " + task["_module"])
		for {
			if !endless && count == 0 {
				break
			}
			count--

			res := m.Perform(p, task)
			if res.Error != nil {
				p.NotifyTelegram("Bot stopping: " + res.Error.Error())
				panic(res.Error)
			}
			fmt.Print(".")

			if !res.CanRepeat {
				break
			}
		}
		fmt.Println()
	}
}
