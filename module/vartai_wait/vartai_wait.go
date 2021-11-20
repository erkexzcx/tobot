package vartai_wait

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"tobot/module"
	"tobot/player"
)

type VartaiWait struct{}

func (obj *VartaiWait) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

var reVartaiTime = regexp.MustCompile(`Vartai atsivers už (\d{2}:\d{2}:\d{2})`)

func (obj *VartaiWait) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kasimas_kalve.php?{{ creds }}&id=deep"

	// Download page
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If levels are too weak - throw error
	foundElements := doc.Find("div:contains('Jūsų damage ir defense vidurkis per žemas, jūs vartams nieko nepadarysite...')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: false, Error: errors.New("damage & defense levels are too low")}
	}

	// If vartai is already open
	foundElements = doc.Find("a:contains('Galite eiti daužyti vartus!')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If vartai does not exist at the moment - find out how much we need to wait and wait
	element := doc.Find("div.antr > b:contains('Pragaro vartų dabar nėra. Vartai atsivers už')")
	if element.Length() > 0 {
		text := strings.TrimSpace(element.Text())
		match := reVartaiTime.FindStringSubmatch(text)
		if len(match) != 2 {
			return &module.Result{CanRepeat: false, Error: errors.New("unable to detect time until vartai opens")}
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
	module.Add("vartai_wait", &VartaiWait{})
}
