package tobot

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"tobot/comms"
	"tobot/module"
	"tobot/player"

	"gopkg.in/yaml.v2"
)

type Activity struct {
	Name  string              `yaml:"name"`
	Tasks []map[string]string `yaml:"tasks"`
}

func Start(p *player.Player) {
	// Create activities from files
	activities := []*Activity{}
	files, err := filepath.Glob(p.Config.ActivitiesDir + string(filepath.Separator) + "*.yml")
	if err != nil {
		log.Fatalln("Failed to read activities .yml files of player '" + p.Config.Nick + "': " + err.Error())
	}
	for _, f := range files {
		if strings.HasPrefix(path.Base(f), "_") {
			continue // Skip '_*.yml' files
		}
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalln(err)
		}

		var a *Activity
		if err := yaml.Unmarshal(contents, &a); err != nil {
			log.Fatalln(err)
		}
		activities = append(activities, a)
	}

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
	comms.SendMessageToTelegram(p.Config.Nick + " started '" + a.Name + "'")

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
				comms.SendMessageToTelegram("Bot stopping: " + res.Error.Error())

				panic(res.Error)
			}

			if !res.CanRepeat {
				break
			}
		}
	}
}
