package vartai

import (
	"errors"
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
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if wrongDoc {
		return obj.Perform(p, settings)
	}

	// If vartai does not exist
	if doc.Find("div:contains('Pragaro vartų dabar nėra!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Smogti!')").Attr("href")
	if !found {
		module.DumpHTML(p, doc)
		return &module.Result{CanRepeat: false, Error: errors.New("action button not found")}
	}

	// Extract request URI from action link
	parsed, err := url.Parse(actionLink)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	requestURI := parsed.RequestURI()

	// Download action page
	doc, wrongDoc, err = p.Navigate("/"+requestURI, true)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if wrongDoc {
		return &module.Result{CanRepeat: true, Error: nil} // No way of knowing if successful, so just assume we can repeat it
	}

	if module.IsInvalidClick(doc) {
		return obj.Perform(p, settings)
	}

	// If action was a success
	if doc.Find("div:contains('Jūs vartams padarote')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If vartai does not exist
	if doc.Find("div:contains('Pragaro vartų dabar nėra!')").Length() > 0 || doc.Find("div:contains('Ir vartus sunaikinate!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(p, doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("vartai", &Vartai{})
}
