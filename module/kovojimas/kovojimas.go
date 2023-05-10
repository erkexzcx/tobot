package kovojimas

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"tobot/module"
	"tobot/module/eating"
	"tobot/player"
)

type Kovojimas struct{}

// Enemy ID and page ID is needed, so this maps enemy to page: ("<enemy>": "page",)
var enemyPage = map[string]string{
	"0":   "0",
	"1":   "0",
	"2":   "0",
	"3":   "0",
	"4":   "0",
	"5":   "0",
	"6":   "0",
	"7":   "1",
	"8":   "1",
	"9":   "1",
	"10":  "1",
	"11":  "1",
	"12":  "1",
	"13":  "1",
	"14":  "1",
	"15":  "1",
	"16":  "1",
	"17":  "1",
	"18":  "1",
	"19":  "1",
	"20":  "2",
	"21":  "2",
	"22":  "2",
	"23":  "2",
	"24":  "2",
	"25":  "2",
	"26":  "2",
	"27":  "3",
	"28":  "3",
	"29":  "3",
	"30":  "3",
	"31":  "3",
	"32":  "3",
	"33":  "3",
	"34":  "3",
	"35":  "4",
	"36":  "4",
	"37":  "4",
	"38":  "4",
	"39":  "4",
	"40":  "4",
	"41":  "4",
	"42":  "4",
	"43":  "4",
	"44":  "5",
	"45":  "5",
	"46":  "5",
	"47":  "5",
	"48":  "5",
	"49":  "5",
	"50":  "5",
	"51":  "5",
	"52":  "6",
	"53":  "6",
	"54":  "6",
	"55":  "6",
	"56":  "6",
	"57":  "6",
	"58":  "7",
	"59":  "7",
	"60":  "7",
	"61":  "7",
	"62":  "7",
	"63":  "7",
	"64":  "7",
	"65":  "7",
	"66":  "7",
	"67":  "7",
	"68":  "7",
	"69":  "8",
	"70":  "8",
	"71":  "8",
	"72":  "8",
	"73":  "8",
	"74":  "8",
	"75":  "8",
	"76":  "9",
	"77":  "9",
	"78":  "9",
	"79":  "9",
	"80":  "9",
	"81":  "9",
	"82":  "9",
	"83":  "9",
	"84":  "9",
	"85":  "9",
	"86":  "10",
	"87":  "10",
	"88":  "10",
	"89":  "10",
	"90":  "10",
	"91":  "10",
	"92":  "10",
	"93":  "10",
	"94":  "10",
	"95":  "10",
	"96":  "10",
	"97":  "10",
	"98":  "10",
	"99":  "10",
	"100": "10",
	"101": "10",
	"102": "10",
	"103": "10",
	"104": "10",
	"105": "10",
	"106": "10",
	"107": "11",
	"108": "11",
	"109": "11",
	"110": "11",
	"111": "11",
	"112": "11",
	"113": "11",
	"114": "11",
	"115": "11",
}

var slayers = map[string]struct{}{
	"1":  {},
	"2":  {},
	"3":  {},
	"4":  {},
	"5":  {},
	"6":  {},
	"7":  {},
	"8":  {},
	"9":  {},
	"10": {},
	"11": {},
	"12": {},
	"13": {},
	"14": {},
	"15": {},
	"16": {},
	"17": {},
}

func (obj *Kovojimas) Validate(settings map[string]string) error {
	// Check if there are any unknown options
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		unknownField := true
		for _, s := range []string{"vs", "slayer", "food", "food_threshold"} {
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
	if _, found := settings["vs"]; !found {
		return errors.New("unrecognized option 'item'")
	}

	// Check if there are any unexpected values
	if _, found := enemyPage[settings["vs"]]; !found {
		return errors.New("unrecognized value of option 'vs'")
	}
	if slayer, found := settings["slayer"]; found {
		if _, found := slayers[slayer]; !found {
			return errors.New("unrecognized value of option 'slayer'")
		}
	}

	// If "food" is set and "food_threshold" is not, consider "food_threshold" equal to 50 (%)
	if item, found := settings["food"]; found {
		if !eating.IsFood(item) {
			return errors.New("unrecognized value of option 'food'")
		}
		if threshold, found := settings["food_threshold"]; found {
			parsed, err := strconv.Atoi(threshold)
			if err != nil {
				return errors.New("unrecognized value of option 'food_threshold' (must be a whole number, from 1 to 100)")
			}
			if parsed < 0 || parsed > 100 {
				return errors.New("unrecognized value of option 'food_threshold' (must be a whole number, from 1 to 100)")
			}
		}
	}

	return nil
}

var reSlayerProgress = regexp.MustCompile(`Atlikta: (\d+) \/ (\d+)`)

func (obj *Kovojimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kova.php?{{ creds }}&id=kova0&vs=" + settings["vs"] + "&psl=" + enemyPage[settings["vs"]]

	// Download page that contains unique action link
	doc, antiCheatPage, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	if antiCheatPage {
		return obj.Perform(p, settings)
	}

	// If slayer is provided - ensure it is enabled
	if slayer, found := settings["slayer"]; found {
		matches := reSlayerProgress.FindStringSubmatch(doc.Text())
		// if slayer not enabled
		if len(matches) != 3 {
			err := enableSlayer(p, slayer)
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			return obj.Perform(p, settings)
		}
		// if slayer completed
		if matches[1] == matches[2] {
			err := finishSlayer(p)
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			return &module.Result{CanRepeat: false, Error: nil}
		}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Pulti')").Attr("href")
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
		// There is no way to extract health bar, so let's assume we need to eat NOW
		if _, found := settings["food"]; found {
			noFoodLeft, err := eating.Eat(p, settings["food"]) // This function goes on loop, so call this once
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			if noFoodLeft {
				return &module.Result{CanRepeat: false, Error: nil}
			}
		}
		return &module.Result{CanRepeat: true, Error: nil}
	}

	if module.IsInvalidClick(doc) {
		return obj.Perform(p, settings)
	}

	if module.IsActionTooFast(doc) {
		return obj.Perform(p, settings)
	}

	// Take some variables
	successWon := doc.Find("div:contains('Nukovėte')").Length() > 0
	successLoss := doc.Find("div:contains('Jūs pralaimėjote')").Length() > 0

	// If fight lost - throw error
	if successLoss {
		return &module.Result{CanRepeat: false, Error: errors.New("fight lost")}
	}

	// If lost - error is already thrown, so the only way - win. If not won - throw error
	if !successWon {
		module.DumpHTML(p, doc)
		return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
	}

	// Check if we can find health bar and reheal accordingly
	if _, found := settings["food"]; found {
		threshold, _ := strconv.Atoi(settings["food_threshold"])
		if threshold == 0 {
			threshold = 50
		}
		_, currentPercent, _, err := eating.ParseHealthPercent(doc.Find("img.hp[src^='graph.php'][src$='c=1']").Eq(1))
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		if currentPercent <= threshold {
			noFoodLeft, err := eating.Eat(p, settings["food"]) // This function goes on loop, so call this once
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

func init() {
	module.Add("kovojimas", &Kovojimas{})
}

func enableSlayer(p *player.Player, slayer string) error {
	path := "/slayer.php?{{ creds }}&id=task&nr=" + slayer

	// Download page that contains unique action link
	doc, antiCheatPage, err := p.Navigate(path, false)
	if err != nil {
		return err
	}
	if antiCheatPage {
		return enableSlayer(p, slayer)
	}

	// Check if we've got a reward for previously completed slayer contract
	if doc.Find("div:contains('Jums sėkmingai paskirta užduotis! Grįžę atgal rasite daugiau informacijos apie užduotį.')").Length() >= 0 {
		return nil
	}

	module.DumpHTML(p, doc)
	return errors.New("enabling slayer contract failed")
}

func finishSlayer(p *player.Player) error {
	path := "/slayer.php?{{ creds }}"

	// Download page that contains unique action link
	doc, antiCheatPage, err := p.Navigate(path, false)
	if err != nil {
		return err
	}
	if antiCheatPage {
		return finishSlayer(p)
	}

	// Check if successfully enabled
	if doc.Find("div:contains('Užduotis atlikta!')").Length() > 0 {
		return nil
	}

	module.DumpHTML(p, doc)
	return errors.New("unable to finish slayer contract")
}
