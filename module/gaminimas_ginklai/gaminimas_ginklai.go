package gaminimas_ginklai

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type GaminimasGinklai struct{}

var items = map[string]struct{}{
	"K1":  {},
	"K2":  {},
	"K3":  {},
	"K4":  {},
	"K5":  {},
	"K6":  {},
	"K7":  {},
	"K8":  {},
	"K9":  {},
	"K10": {},
	"K11": {},
	"K12": {},
	"K13": {},
	"K14": {},
	"K15": {},
	"K16": {},
	"K17": {},
	"K18": {},
	"K19": {},
	"K20": {},
	"K21": {},
	"K22": {},
	"K23": {},
	"K24": {},
	"K25": {},
	"K26": {},
	"K27": {},
	"K28": {},
	"K29": {},
	"K30": {},
	"K31": {},
	"K32": {},
	"K33": {},
	"K34": {},
	"K37": {},
	"K38": {},
	"K40": {},
}

func (obj *GaminimasGinklai) Validate(settings map[string]string) error {
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

func (obj *GaminimasGinklai) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kasimas_kalve.php?{{ creds }}&id=kaldinti2&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, antiCheatPage, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if antiCheatPage {
		return obj.Perform(p, settings)
	}

	// Check if not depleted
	if doc.Find("b:contains('Nepakanka žaliavų!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Kalti')").Attr("href")
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
	doc, antiCheatPage, err = p.Navigate("/"+requestURI, true)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if antiCheatPage {
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
	if doc.Find("div:contains('Nukalta: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(p, doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("gaminimas_ginklai", &GaminimasGinklai{})
}
