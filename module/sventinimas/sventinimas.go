package sventinimas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Sventinimas struct{}

func (obj *Sventinimas) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

func (obj *Sventinimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/baznycia.php?{{ creds }}"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If don't have enough resources
	if doc.Find("div:contains('Neturite vandens!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Šventinti vandenį')").Attr("href")
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
	if doc.Find("div:contains('Vanduo pašventintas')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If don't have enough resources
	if doc.Find("div:contains('Neturite vandens!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("sventinimas", &Sventinimas{})
}
