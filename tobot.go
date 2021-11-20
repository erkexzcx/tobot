package tobot

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
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
	log.Print("Started '" + a.Name + "'")

	for _, task := range a.Tasks {
		m := module.Modules[task["_module"]]

		count, _ := strconv.Atoi(task["_count"])
		endless := false
		if count == 0 {
			endless = true
		}

		log.Println(moduleSettingsToTitle(task))
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
			fmt.Print(".")

			if !res.CanRepeat {
				break
			}
		}
		fmt.Println()
	}
}

func moduleSettingsToTitle(m map[string]string) string {
	list := []string{}
	for k, v := range m {
		if k == "_module" {
			continue
		}
		list = append(list, k+":"+v)
	}
	sort.StringsAreSorted(list)
	return "Module: " + m["_module"] + "{" + strings.Join(list, "; ") + "}"
}
