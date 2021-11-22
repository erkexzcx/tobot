package gaminimas_ginklai

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type GaminimasGinklai struct{}

var allowedSettings = map[string][]string{
	"item": {
		"K1",
		"K2",
		"K3",
		"K4",
		"K5",
		"K6",
		"K7",
		"K8",
		"K9",
		"K10",
		"K11",
		"K12",
		"K13",
		"K14",
		"K15",
		"K16",
		"K17",
		"K18",
		"K19",
		"K20",
		"K21",
		"K22",
		"K23",
		"K24",
		"K25",
		"K26",
		"K27",
		"K28",
		"K29",
		"K30",
	},
}

func (obj *GaminimasGinklai) Validate(settings map[string]string) error {
	// Check for missing keys
	for k := range allowedSettings {
		_, found := settings[k]
		if !found {
			return errors.New("missing key '" + k + "'")
		}
	}

	for k, v := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}

		// Check for unknown keys
		_, found := allowedSettings[k]
		if !found {
			return errors.New("unrecognized key '" + k + "'")
		}

		// Check for unknown value
		found = false
		for _, el := range allowedSettings[k] {
			if el == v {
				found = true
				break
			}
		}
		if !found {
			return errors.New("unrecognized value of key '" + k + "'")
		}
	}

	return nil
}

func (obj *GaminimasGinklai) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kasimas_kalve.php?{{ creds }}&id=kaldinti2&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Check if not depleted
	if doc.Find("b:contains('Nepakanka žaliavų!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Kalti')").Attr("href")
	if !found {
		module.DumpHTML(doc)
		return &module.Result{CanRepeat: false, Error: errors.New("action button not found")}
	}

	// Extract request URI from action link
	parsed, err := url.Parse(actionLink)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	requestURI := parsed.RequestURI()

	// Download action page
	doc, err = p.Navigate("/"+requestURI, true)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	if doc.Find("div:contains('Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!')").Length() > 0 {
		return obj.Perform(p, settings)
	}

	// If action was a success
	if doc.Find("div:contains('Nukalta: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If actioned too fast
	if doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length() > 0 {
		log.Println("actioned too fast, retrying...")
		return obj.Perform(p, settings)
	}

	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("gaminimas_ginklai", &GaminimasGinklai{})
}
