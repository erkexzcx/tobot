package trolis

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/module/eating"
	"tobot/player"
)

type Trolis struct{}

func (obj *Trolis) Validate(settings map[string]string) error {
	food, found := settings["eating"]
	if !found {
		return errors.New("missing 'eating' field")
	}
	if !eating.IsEatable(food) {
		return errors.New("unknown 'eating' field")
	}

	for k := range settings {
		if strings.HasPrefix(k, "_") || k == "eating" {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}

	return nil
}

func (obj *Trolis) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/zaisti.php?{{ creds }}&id=fighttroll"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Smogti troliui')").Attr("href")
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
	if doc.Find("div:contains('Padaryta žala:')").Length() > 0 {
		if _, found := settings["eating"]; found {
			currentHealth, _, _, err := eating.ParseHealthPercent(doc.Find("img.hp[src^='graph.php'][src$='c=1']"))
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			if currentHealth == 0 {
				noFoodLeft, err := eating.Eat(p, settings["eating"]) // This function goes on loop, so call this once
				if err != nil {
					return &module.Result{CanRepeat: false, Error: err}
				}
				if noFoodLeft {
					return &module.Result{CanRepeat: false, Error: nil}
				}
			}
		}
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If troll does not exist
	if doc.Find("div:contains('Trolio nematyt...')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("trolis", &Trolis{})
}
