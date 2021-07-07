package kirtimas

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Kirtimas struct{}

var allowedSettings = map[string][]string{
	"item": {
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
	},
}

func (obj *Kirtimas) Validate(settings map[string]string) error {
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

func (obj *Kirtimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/miskas.php?{{ creds }}&id=kertu&ka=" + settings["item"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Kirsti')").Attr("href")
	if !found {
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

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	foundElements := doc.Find("div:contains('Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!')").Length()
	if foundElements > 0 {
		return obj.Perform(p, settings)
	}

	// If action was a success
	foundElements = doc.Find("div:contains('Nukirsta: ')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If inventory full
	foundElements = doc.Find("div:contains('Jūsų inventorius jau pilnas!')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If actioned too fast
	foundElements = doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length()
	if foundElements > 0 {
		log.Println("actioned too fast, retrying...")
		return obj.Perform(p, settings)
	}

	html, _ := doc.Html()
	log.Println(html)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("kirtimas", &Kirtimas{})
}
