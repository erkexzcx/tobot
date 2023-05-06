package apdirbimas_dragon

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type ApdirbimasDragon struct{}

func (obj *ApdirbimasDragon) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		unknownField := true
		if unknownField {
			return errors.New("unrecognized option '" + k + "'")
		}
	}

	return nil
}

func (obj *ApdirbimasDragon) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/dirbtuves.php?{{ creds }}&id="

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Apdirbti dragon akmenį')").Attr("href")
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

	// Ignore if no more akmenys
	if doc.Find(":contains('Nepakanka dragon akmenų')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Ignore if level too low
	if doc.Find("div:contains('lygis per žemas')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If action was a success
	if doc.Find("div:contains('Akmuo apdirbtas')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("apdirbimas_dragon", &ApdirbimasDragon{})
}
