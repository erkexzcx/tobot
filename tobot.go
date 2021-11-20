package tobot

import (
	"strconv"
	"tobot/module"
	"tobot/player"
)

type Activity struct {
	Name  string              `yaml:"name"`
	Tasks []map[string]string `yaml:"tasks"`
}

func Start(p *player.Player, activities []*Activity) {
	// Validate activities
	for _, a := range activities {
		validateActivity(a)
	}

	// Run activities in a loop
	for {
		for _, aa := range activities {
			runActivity(p, aa)
		}
	}
}

func validateActivity(a *Activity) {
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

func runActivity(p *player.Player, a *Activity) {
	p.Println("Started '" + a.Name + "'")

	for _, task := range a.Tasks {
		m := module.Modules[task["_module"]]

		count, _ := strconv.Atoi(task["_count"])
		endless := false
		if count == 0 {
			endless = true
		}

		for {
			if !endless && count == 0 {
				break
			}
			count--

			res := m.Perform(p, task)
			if res.Error != nil {
				p.NotifyTelegram("Bot stopping: "+res.Error.Error(), false)
				panic(res.Error)
			}

			if !res.CanRepeat {
				break
			}
		}
	}
}
