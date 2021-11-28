package zvejyba

import (
	"errors"
	"net/url"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Zvejyba struct{}

var items = map[string]string{
	"sliekas": "/zvejoti.php?{{ creds }}",
	"tesla":   "/zvejoti.php?{{ creds }}",
	"karos":   "/zvejoti.php?{{ creds }}",
	"zui":     "/zvejoti.php?{{ creds }}",
	"el":      "/zvejoti.php?{{ creds }}&id=jura",
	"tink":    "/zvejoti.php?{{ creds }}&id=jura",
	"zeb":     "/zvejoti.php?{{ creds }}&id=jura",
	"biz":     "/zvejoti.php?{{ creds }}&id=jura",
	"t1":      "/zvejoti.php?{{ creds }}&id=vandenynas",
	"t2":      "/zvejoti.php?{{ creds }}&id=vandenynas",
	"t3":      "/zvejoti.php?{{ creds }}&id=vandenynas",
	"t4":      "/zvejoti.php?{{ creds }}&id=vandenynas",
	"t5":      "/zvejoti.php?{{ creds }}&id=vandenynas",
}

func (obj *Zvejyba) Validate(settings map[string]string) error {
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

func (obj *Zvejyba) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := items[settings["item"]]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd='][href*='ka=" + settings["item"] + "']").Attr("href")
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

	if doc.Find("div:contains('Nepakanka')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// If action was a success
	if doc.Find("div:contains('Pagavote ')").Length()+doc.Find("div:contains('Nieko nepagavote')").Length() > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If inventory full
	if doc.Find("div:contains('Jūsų inventorius jau pilnas!')").Length() > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}
	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("zvejyba", &Zvejyba{})
}
