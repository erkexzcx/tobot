package eating

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"tobot/module"
	"tobot/player"

	"github.com/PuerkitoBio/goquery"
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

var reGyvybes = regexp.MustCompile(`Gyvybės: (\d+\.?\d*)\/(\d+)`)

func (obj *Eating) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		for _, s := range []string{"food"} {
			if k == s {
				continue
			}
			return errors.New("unrecognized option '" + k + "'")
		}
	}

	// Check if any mandatory option is missing
	if _, found := settings["food"]; !found {
		return errors.New("missing option 'food'")
	}

	// Check if there are any unexpected values
	if !IsFood(settings["food"]) {
		return errors.New("unrecognized value of option 'food'")
	}

	return nil
}

func (obj *Eating) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/namai.php?{{ creds }}&id=lova"

	// Download page that contains info about health
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if wrongDoc {
		return obj.Perform(p, settings)
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

func IsFood(item string) bool {
	if _, found := eatables[item]; found {
		return true
	}
	return false
}

func Eat(p *player.Player, item string) (noFoodLeft bool, err error) {
	path := "/zaisti.php?{{ creds }}&id=valgyti&ka=" + item

	// Download page
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return false, err
	}
	if wrongDoc {
		return Eat(p, item) // Eat again, since we don't know what happened
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

	module.DumpHTML(p, doc)
	return false, errors.New("unknown error occurred (at eating module)")
}

// Extracts current health and max health out of player health bar. Efficient method
// to find out how much health left after fight/hit
func ParseHealthPercent(healthBar *goquery.Selection) (remainingHealth, remainingPercent, maxHealth int, err error) {
	val, found := healthBar.Attr("src")
	if !found {
		return 0, 0, 0, errors.New("health bar or it's source not found")
	}
	val = strings.ReplaceAll(val, "graph.php?", "")
	valPairs := strings.Split(val, "&") // Yes, this is not &amp;
	for _, v := range valPairs {
		valPairParts := strings.Split(v, "=")
		if valPairParts[0] == "iki" {
			maxHealth, _ = strconv.Atoi(valPairParts[1])
		}
		if valPairParts[0] == "yra" {
			tmp := strings.Split(valPairParts[1], ".")
			remainingHealth, _ = strconv.Atoi(tmp[0])
		}
	}
	if maxHealth == 0 {
		return 0, 0, 0, errors.New("failed to read health bar values")
	}
	return remainingHealth, remainingPercent, int(float64(remainingHealth) / float64(maxHealth) * 100), nil
}
