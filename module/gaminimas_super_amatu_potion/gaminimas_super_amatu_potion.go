package gaminimas_super_amatu_potion

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type GaminimasSuperAmatuPotion struct{}

// In shop items found as 'SPA1' instead of '1'
var items = map[string]struct{}{
	"1":  {},
	"2":  {},
	"3":  {},
	"4":  {},
	"5":  {},
	"6":  {},
	"7":  {},
	"8":  {},
	"9":  {},
	"10": {},
	"11": {},
	"12": {},
}

func (obj *GaminimasSuperAmatuPotion) Validate(settings map[string]string) error {
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

func (obj *GaminimasSuperAmatuPotion) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/namai.php?{{ creds }}&id=samatupotion02&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Gaminti')").Attr("href")
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

	// Ignore if level too low
	if doc.Find("div:contains('lygis per žemas')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If action was a success
	if doc.Find("div:contains('Pagaminta: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// Check if not depleted
	if doc.Find(":contains('Nepakanka reikiamų uogų!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("gaminimas_super_amatu_potion", &GaminimasSuperAmatuPotion{})
}
