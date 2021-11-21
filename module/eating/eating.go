package eating

import (
	"errors"
	"regexp"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Eating struct{}

var eatables = map[string]struct{}{
	"MK1":   {},
	"KZ1":   {},
	"UO1":   {},
	"KGR1":  {},
	"MK2":   {},
	"KZ2":   {},
	"UO2":   {},
	"KGR2":  {},
	"MK3":   {},
	"KZ3":   {},
	"UO3":   {},
	"KGR3":  {},
	"MK4":   {},
	"KZ4":   {},
	"UO4":   {},
	"KGR4":  {},
	"MK5":   {},
	"KZ5":   {},
	"UO5":   {},
	"KGR5":  {},
	"MK6":   {},
	"KZ6":   {},
	"UO6":   {},
	"KGR6":  {},
	"MK7":   {},
	"KZ7":   {},
	"UO7":   {},
	"KGR7":  {},
	"MK8":   {},
	"KZ8":   {},
	"UO8":   {},
	"KGR8":  {},
	"MK9":   {},
	"KZ9":   {},
	"UO9":   {},
	"KGR9":  {},
	"MK10":  {},
	"KZ10":  {},
	"UO10":  {},
	"KGR10": {},
	"MK11":  {},
	"KZ11":  {},
	"UO11":  {},
	"KGR11": {},
	"MK12":  {},
	"KZ12":  {},
	"UO12":  {},
	"KGR12": {},
	"MK13":  {},
	"KZ13":  {},
	"UO13":  {},
	"KGR13": {},
	"MK14":  {},
	"KZ14":  {},
	"UO14":  {},
	"MK15":  {},
	"KZ15":  {},
	"MK16":  {},
	"KZ16":  {},
	"MK17":  {},
	"KZ17":  {},
	"MK18":  {},
	"KZ18":  {},
	"MK19":  {},
	"KZ19":  {},
	"KZ20":  {},
}

var reGyvybes = regexp.MustCompile(`Gyvybės: (\d+)\/(\d+)`)

func (obj *Eating) Validate(settings map[string]string) error {
	food, found := settings["food"]
	if !found {
		return errors.New("missing 'food' field")
	}
	if !IsEatable(food) {
		return errors.New("unknown 'food' field")
	}

	for k := range settings {
		if strings.HasPrefix(k, "_") || k == "food" {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}

	return nil
}

func (obj *Eating) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/namai.php?{{ creds }}&id=lova"

	// Download page that contains info about health
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Extract health and eat if not full
	docText := doc.Text()
	match := reGyvybes.FindStringSubmatch(docText)
	if len(match) != 3 {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to find gyvybes regex match")}
	}
	if match[1] == match[2] {
		return &module.Result{CanRepeat: false, Error: nil}
	}
	_, err = Eat(p, settings["food"]) // This loops as long as it takes and then returns value
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	return &module.Result{CanRepeat: false, Error: nil}
}

func init() {
	module.Add("eating", &Eating{})
}

func IsEatable(item string) bool {
	if _, found := eatables[item]; found {
		return true
	}
	return false
}

func Eat(p *player.Player, item string) (noFoodLeft bool, err error) {
	path := "/zaisti.php?{{ creds }}&id=valgyti&ka=" + item

	// Download page
	doc, err := p.Navigate(path, false)
	if err != nil {
		return false, err
	}

	// Check if ran out of food
	if doc.Find("div:contains('Šio maisto neturite!')").Length() > 0 {
		return true, nil
	}

	// Check if food was eaten successfully
	if doc.Find("div:contains('Suvalgyta')").Length() > 0 {
		docText := doc.Text()
		match := reGyvybes.FindStringSubmatch(docText)
		if len(match) != 3 {
			return false, errors.New("unable to find regex match (after eathing health)")
		}
		if match[1] == match[2] {
			return false, nil
		}
		return Eat(p, item)
	}

	module.DumpHTML(doc)
	return false, errors.New("unknown error occurred (at eating module)")
}
