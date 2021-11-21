package demonas_wait

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"tobot/module"
	"tobot/player"
)

type DemonasWait struct{}

func (obj *DemonasWait) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

var reDemonasTime = regexp.MustCompile(`Demonas prisikels už: (\d{2}:\d{2}:\d{2})`)

func (obj *DemonasWait) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kova.php?{{ creds }}&id=tobgod"

	// Download page
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// demonas is available
	if doc.Find("a[href*='&kd=']:contains('Eiti į kovą su tob demonu!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If demonas does not exist at the moment - find out how much we need to wait and wait
	if doc.Find("div.antr > b:contains('Demonas prisikels už:')").Length() > 0 {
		match := reDemonasTime.FindStringSubmatch(doc.Text())
		if len(match) != 2 {
			return &module.Result{CanRepeat: false, Error: errors.New("unable to detect time until demonas spawns")}
		}

		durationParts := strings.Split(match[1], ":")
		durationString := durationParts[0] + "h" + durationParts[1] + "m" + durationParts[2] + "s"
		duration, _ := time.ParseDuration(durationString)
		time.Sleep(duration)
		return &module.Result{CanRepeat: false, Error: nil}
	}

	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("demonas_wait", &DemonasWait{})
}
