package gaminimas_batai

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type GaminimasBatai struct{}

var items = map[string]struct{}{
	"BA1":  {},
	"BA2":  {},
	"BA3":  {},
	"BA4":  {},
	"BA5":  {},
	"BA6":  {},
	"BA7":  {},
	"BA8":  {},
	"BA9":  {},
	"BA10": {},
	"BA11": {},
	"BA12": {},
	"BA13": {},
}

func (obj *GaminimasBatai) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		unknownField := true
		for _, s := range []string{"item"} {
			if k == s {
				unknownField = false
				break
			}
		}
		if unknownField {
			return errors.New("unrecognized option '" + k + "'")
		}
	}

	// Check if any mandatory option is missing
	if _, found := settings["item"]; !found {
		return errors.New("unrecognized option 'item'")
	}

	// Check if there are any unexpected values
	if _, found := items[settings["item"]]; !found {
		return errors.New("unrecognized value of option 'item'")
	}

	return nil
}

func (obj *GaminimasBatai) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/dirbtuves.php?{{ creds }}&id=fmat0&ka=" + settings["item"] + "&page=3"

	// Download page that contains unique action link
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if wrongDoc {
		return obj.Perform(p, settings)
	}

	// Check if not depleted
	if doc.Find("b:contains('Nepakanka žaliavų!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Gaminti')").Attr("href")
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
		return &module.Result{CanRepeat: true, Error: nil} // There is no way to know if action was successful, so just assume it was
	}

	if module.IsInvalidClick(doc) {
		return obj.Perform(p, settings)
	}

	// Ignore if level too low
	if doc.Find("div:contains('lygis per žemas')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If action was a success
	if doc.Find("div:contains('Pagaminta:')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(p, doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("gaminimas_batai", &GaminimasBatai{})
}
