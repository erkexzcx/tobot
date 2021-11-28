package kepimas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Kepimas struct{}

var allowedSettings = map[string][]string{
	"item": {
		// Nekeptos zuvys
		"Z1",
		"Z2",
		"Z3",
		"Z4",
		"Z5",
		"Z6",
		"Z7",
		"Z8",
		"Z9",
		"Z10",
		"Z11",
		"Z12",
		"Z13",
		"Z14",
		"Z15",
		"Z16",

		// Nekepta mesa
		"MS1",
		"MS2",
		"MS3",
		"MS4",
		"MS5",
		"MS6",
		"MS7",
		"MS8",
		"MS9",
		"MS10",
		"MS11",
		"MS12",
		"MS13",

		// Nekepti grybai
		"GR1",
		"GR2",
		"GR3",
		"GR4",
		"GR5",
		"GR6",
		"GR7",
		"GR8",
		"GR9",
		"GR10",
		"GR11",
		"GR12",
	},
	"fuel": {
		// Mediena
		"MA1",
		"MA2",
		"MA3",
		"MA4",
		"MA5",
		"MA6",
		"MA7",
		"MA8",
		"MA9",
		"MA10",
		"MA11",
		"MA12",
		"MA13",
		"MA14",
		"MA15",
		"MA16",

		// Anglis
		"O6",
	},
}

func (obj *Kepimas) Validate(settings map[string]string) error {
	// Check for missing keys
	for k := range allowedSettings {
		_, found := settings[k]
		if !found {
			return errors.New("missing key '" + k + "'")
		}
	}

	for k, v := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}

		// Check for unknown keys
		_, found := allowedSettings[k]
		if !found {
			return errors.New("unrecognized key '" + k + "'")
		}

		// Check for unknown value
		found = false
		for _, el := range allowedSettings[k] {
			if el == v {
				found = true
				break
			}
		}
		if !found {
			return errors.New("unrecognized value of key '" + k + "'")
		}
	}

	return nil
}

func (obj *Kepimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/namai.php?{{ creds }}&id=gaminu2&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If need fuel
	if doc.Find("div:contains('Ugnis užgeso...!')").Length() > 0 {
		err := addFuel(p, settings)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		return obj.Perform(p, settings)
	}

	// Check if not depleted
	if doc.Find("b:contains('Nebeturite ko kepti!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Kepti')").Attr("href")
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

	// If need fuel
	if doc.Find("div:contains('Ugnis užgeso...!')").Length() > 0 {
		err := addFuel(p, settings)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		return obj.Perform(p, settings)
	}

	// If action was a success
	if doc.Find("div:contains(' (jau turite: ')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func addFuel(p *player.Player, settings map[string]string) error {
	path := "/namai.php?{{ creds }}&id=kurt&ka=" + settings["fuel"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, true)
	if err != nil {
		return err
	}

	// If action was a success
	if doc.Find("div:contains('Krosnelė užkurta, galite eiti kepti maistą.')").Length() > 0 {
		return nil
	}

	return errors.New("failed to add fuel to krosnele")
}

func init() {
	module.Add("kepimas", &Kepimas{})
}
