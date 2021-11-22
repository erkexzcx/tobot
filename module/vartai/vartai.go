package vartai

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Vartai struct{}

func (obj *Vartai) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

func (obj *Vartai) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kasimas_kalve.php?{{ creds }}&id=fightgate"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If vartai does not exist
	if doc.Find("div:contains('Pragaro vartų dabar nėra!')").Length() > 0 {
		return obj.Perform(p, settings)
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Smogti!')").Attr("href")
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
	if doc.Find("div:contains('Jūs vartams padarote')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If vartai does not exist
	if doc.Find("div:contains('Pragaro vartų dabar nėra!')").Length() > 0 {
		return obj.Perform(p, settings)
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
	module.Add("vartai", &Vartai{})
}
