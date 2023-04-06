package kepimas

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Kepimas struct{}

var items = map[string]struct{}{
	"Z1":   {},
	"Z2":   {},
	"Z3":   {},
	"Z4":   {},
	"Z5":   {},
	"Z6":   {},
	"Z7":   {},
	"Z8":   {},
	"Z9":   {},
	"Z10":  {},
	"Z11":  {},
	"Z12":  {},
	"Z13":  {},
	"Z14":  {},
	"Z15":  {},
	"Z16":  {},
	"MS1":  {},
	"MS2":  {},
	"MS3":  {},
	"MS4":  {},
	"MS5":  {},
	"MS6":  {},
	"MS7":  {},
	"MS8":  {},
	"MS9":  {},
	"MS10": {},
	"MS11": {},
	"MS12": {},
	"MS13": {},
	"GR1":  {},
	"GR2":  {},
	"GR3":  {},
	"GR4":  {},
	"GR5":  {},
	"GR6":  {},
	"GR7":  {},
	"GR8":  {},
	"GR9":  {},
	"GR10": {},
	"GR11": {},
	"GR12": {},
}

var fuels = map[string]struct{}{
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
	"O6":   {},
}

func (obj *Kepimas) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		unknownField := true
		for _, s := range []string{"item", "fuel"} {
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
	if _, found := settings["fuel"]; !found {
		return errors.New("unrecognized option 'fuel'")
	}

	// Check if there are any unexpected values
	if _, found := items[settings["item"]]; !found {
		return errors.New("unrecognized value of option 'item'")
	}
	if _, found := fuels[settings["fuel"]]; !found {
		return errors.New("unrecognized value of option 'fuel'")
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

	// Ignore if level too low
	if doc.Find("div:contains('lygis per žemas')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
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
