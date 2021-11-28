package apdirbimas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Apdirbimas struct{}

var items = map[string]struct{}{
	"AB1":  {},
	"AB2":  {},
	"AB3":  {},
	"AB4":  {},
	"AB5":  {},
	"AB6":  {},
	"AB7":  {},
	"AB8":  {},
	"AB9":  {},
	"AB10": {},
	"AB11": {},
	"AB12": {},
	"AB13": {},
	"AB14": {},
}

func (obj *Apdirbimas) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		for _, s := range []string{"item"} {
			if k == s {
				continue
			}
		}
		return errors.New("unrecognized option '" + k + "'")
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

func (obj *Apdirbimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/dirbtuves.php?{{ creds }}&id=bapd0&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Check if not depleted
	if doc.Find("b:contains('Neužtenka neapdirbtų akmenų!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Apdirbti')").Attr("href")
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
	if doc.Find("div:contains('Apdirbta: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("apdirbimas", &Apdirbimas{})
}
