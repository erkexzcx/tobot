package demonas

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/module/all/eating"
	"tobot/player"
)

type Demonas struct{}

func (obj *Demonas) Validate(settings map[string]string) error {
	for k, v := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if k == "eating" {
			if !eating.IsEatable(v) {
				return errors.New("unrecognized value of key '" + k + "'")
			}
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
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

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	foundElements := doc.Find("div:contains('Taip negalima! Turite eiti atgal ir vėl pulti!')").Length()
	if foundElements > 0 {
		return obj.Perform(p, settings)
	}

	// If action was a success
	res := doc.Find("div:contains('Jūs demonui nuimate gyvybių:')").Length() > 0 ||
		doc.Find("div:contains('Sužalotas negalite kautis prieš demoną. Gyvybės turi būti pilnos.')").Length() > 0 ||
		doc.Find("div:contains('Pasipildykite gyvybes.')").Length() > 0
	if res {
		outOfFood, err := eating.Eat(p, settings["eating"])
		if err != nil {
			return &module.Result{CanRepeat: true, Error: err}
		}
		if outOfFood {
			return &module.Result{CanRepeat: false, Error: nil}
		}
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If actioned too fast
	foundElements = doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length()
	if foundElements > 0 {
		log.Println("actioned too fast, retrying...")
		return obj.Perform(p, settings)
	}

	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("demonas", &Demonas{})
}
