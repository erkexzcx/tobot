package kirtimas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Kirtimas struct{}

var items = map[string]struct{}{
	"MA1":  {},
	"MA2":  {},
	"MA3":  {},
	"MA4":  {},
	"MA5":  {},
	"MA6":  {},
	"MA7":  {},
	"MA8":  {},
	"MA9":  {},
	"MA10": {},
	"MA11": {},
	"MA12": {},
	"MA13": {},
	"MA14": {},
	"MA15": {},
	"MA16": {},
	"MA17": {},
	"MA18": {},
	"MA19": {},
	"MA20": {},
	"MA21": {},
}

func (obj *Kirtimas) Validate(settings map[string]string) error {
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

func (obj *Kirtimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/miskas.php?{{ creds }}&id=kertu&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if wrongDoc {
		return obj.Perform(p, settings)
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Kirsti')").Attr("href")
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

	// If action was a success
	if doc.Find("div:contains('Nukirsta: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If inventory full
	if doc.Find("div:contains('Jūsų inventorius jau pilnas!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(p, doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("kirtimas", &Kirtimas{})
}
