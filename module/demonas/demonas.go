package demonas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/module/eating"
	"tobot/player"
)

type Demonas struct{}

func (obj *Demonas) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		for _, s := range []string{"food"} {
			if k == s {
				continue
			}
		}
		return errors.New("unrecognized option '" + k + "'")
	}

	// Check if any mandatory option is missing
	if _, found := settings["food"]; !found {
		return errors.New("unrecognized option 'food'")
	}

	// Check if there are any unexpected values
	if !eating.IsFood(settings["food"]) {
		return errors.New("unrecognized value of option 'food'")
	}

	return nil
}

func (obj *Demonas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kova.php?{{ creds }}&id=tobgod"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Demon is still dead
	if doc.Find("div:contains('Demonas prisikels už')").Length() > 0 {
		return obj.Perform(p, settings)
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Eiti į kovą su tob demonu!')").Attr("href")
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

	if module.IsInvalidClick(doc) {
		return obj.Perform(p, settings)
	}

	// If action was a success
	res := doc.Find("div:contains('Jūs demonui nuimate gyvybių:')").Length() > 0 ||
		doc.Find("div:contains('Sužalotas negalite kautis prieš demoną. Gyvybės turi būti pilnos.')").Length() > 0 ||
		doc.Find("div:contains('Pasipildykite gyvybes.')").Length() > 0
	if res {
		outOfFood, err := eating.Eat(p, settings["food"])
		if err != nil {
			return &module.Result{CanRepeat: true, Error: err}
		}
		if outOfFood {
			return &module.Result{CanRepeat: false, Error: nil}
		}
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("demonas", &Demonas{})
}
