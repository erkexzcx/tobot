package tobot

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"tobot/comms"
	"tobot/module"
	"tobot/player"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v3"
)

var log = logging.MustGetLogger("global")

type Activity struct {
	Name  string              `yaml:"name"`
	Tasks []map[string]string `yaml:"tasks"`
}

func Start(p *player.Player) {
	// Create activities from files
	activities := []*Activity{}
	directoriesLocation := p.Config.ActivitiesDir + string(filepath.Separator) + "*.yml"
	files, err := filepath.Glob(directoriesLocation)
	if err != nil {
		p.Log.Critical("Failed to read activities .yml files of player '" + p.Config.Nick + "': " + err.Error())
	}

	p.Log.Debug("Parsing activity files from ", directoriesLocation)
	for _, f := range files {
		p.Log.Debug("Processing file:", f)
		if strings.HasPrefix(path.Base(f), "_") {
			p.Log.Debug("Skipping activity file:", f)
			continue // Skip '_*.yml' files
		}
		contents, err := os.ReadFile(f)
		if err != nil {
			p.Log.Criticalf("Failed to read activity file %s:%s", f, err.Error())
		}

		var a *Activity
		if err := yaml.Unmarshal(contents, &a); err != nil {
			p.Log.Criticalf("Failed to unmarshal activity file %s:%s", f, err.Error())
		}
		activities = append(activities, a)
		p.Log.Debug("Activity loaded:", a.Name)
	}

	// Validate activities
	for _, a := range activities {
		validateActivity(a)
	}
	p.Log.Debug("Activities validated successfully")

	// Run activities in a loop
	p.Log.Info("Activities started!")
	for {
		for _, aa := range activities {
			p.Log.Infof("Starting activity '%s'", aa.Name)
			runActivity(p, aa)
		}
	}
}

func validateActivity(a *Activity) {
	for _, task := range a.Tasks {
		count, found := task["_count"]
		if found {
			countInt, err := strconv.Atoi(count)
			if err != nil || countInt < 0 {
				log.Panicf("invalid '_count' value (in '%s' activity)\n", a.Name)
			}
		}
		if task["_module"] == "" {
			log.Panicf("task missing _module (in '%s' activity)\n", a.Name)
		}
		m, found := module.Modules[task["_module"]]
		if !found {
			log.Panicf("unknown _module '%s' (in '%s' activity)\n", task["_module"], a.Name)
		}
		err := m.Validate(task)
		if err != nil {
			log.Panicf("Error from activity '%s' module '%s': %s\n", a.Name, task["_module"], err.Error())
		}
	}
}

func runActivity(p *player.Player, a *Activity) {
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
				comms.SendMessageToTelegram("Bot stopped: " + res.Error.Error())
				p.Log.Critical("Bot stopped:", res.Error)
				select {}
			}

			if !res.CanRepeat {
				break
			}
		}
	}
}
