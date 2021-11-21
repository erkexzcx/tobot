package kovojimas

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"
	"tobot/module"
	"tobot/module/eating"
	"tobot/player"
)

type Kovojimas struct{}

var enemyToPageMap = map[string]string{
	// Požemis
	"0": "0",
	"1": "0",
	"2": "0",
	"3": "0",
	"4": "0",
	"5": "0",
	"6": "0",

	// Miškas
	"7":  "1",
	"8":  "1",
	"9":  "1",
	"10": "1",
	"11": "1",
	"12": "1",
	"13": "1",
	"14": "1",
	"15": "1",
	"16": "1",
	"17": "1",
	"18": "1",
	"19": "1",

	// Dykuma
	"20": "2",
	"21": "2",
	"22": "2",
	"23": "2",
	"24": "2",
	"25": "2",
	"26": "2",

	// Užburtas miškas
	"27": "3",
	"28": "3",
	"29": "3",
	"30": "3",
	"31": "3",
	"32": "3",
	"33": "3",
	"34": "3",

	// Užburtas kraštas
	"35": "4",
	"36": "4",
	"37": "4",
	"38": "4",
	"39": "4",
	"40": "4",
	"41": "4",
	"42": "4",
	"43": "4",

	// Išmiręs miestelis
	"44": "5",
	"45": "5",
	"46": "5",
	"47": "5",
	"48": "5",
	"49": "5",
	"50": "5",
	"51": "5",

	// Apleistas namas
	"52": "6",
	"53": "6",
	"54": "6",
	"55": "6",
	"56": "6",
	"57": "6",

	// Drakonų urvai
	"58": "7",
	"59": "7",
	"60": "7",
	"61": "7",
	"62": "7",
	"63": "7",
	"64": "7",
	"65": "7",
	"66": "7",
	"67": "7",
	"68": "7",

	// Ugnies žemė
	"69": "8",
	"70": "8",
	"71": "8",
	"72": "8",
	"73": "8",
	"74": "8",
	"75": "8",

	// Sniegynas
	"76": "9",
	"77": "9",
	"78": "9",
	"79": "9",
	"80": "9",
	"81": "9",
	"82": "9",
	"83": "9",
	"84": "9",
	"85": "9",

	// Mirties sala
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
	"107": "10",
	"108": "10",
	"109": "10",
	"110": "10",
	"111": "10",
	"112": "10",
	"113": "10",
	"114": "10",
	"115": "10",
}

var allowedSlayers = map[string]struct{}{
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
	enemy, found := settings["vs"]
	if !found {
		return errors.New("missing 'vs' field")
	}
	_, found = enemyToPageMap[enemy]
	if !found {
		return errors.New("unknown 'vs' field")
	}

	slayer, found := settings["slayer"]
	if found {
		_, found = allowedSlayers[slayer]
		if !found {
			return errors.New("unknown value of 'slayer' field")
		}
	}

	// If "eating" is set and "eating_threshold" is not, consider "eating_threshold" equal to 50 (%)
	if item, found := settings["eating"]; found {
		if !eating.IsEatable(item) {
			return errors.New("provided value of field 'eating' is not a food")
		}
		if threshold, found := settings["eating_threshold"]; found {
			parsed, err := strconv.Atoi(threshold)
			if err != nil {
				return errors.New("provided value of field 'eating_threshold' is not a number (must contain only digits)")
			}
			if parsed < 0 || parsed > 100 {
				return errors.New("provided value of field 'eating_threshold' can only be within a range from 1 to 100")
			}
		}
	}

	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if k != "vs" && k != "eating" && k != "eating_threshold" && k != "slayer" {
			return errors.New("unknown '" + k + "' field")
		}
	}

	return nil
}

func (obj *Kovojimas) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kova.php?{{ creds }}&id=kova0&vs=" + settings["vs"] + "&psl=" + enemyToPageMap[settings["vs"]]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If slayer is provided - ensure it is enabled
	if slayer, found := settings["slayer"]; found {
		slayerInProgress := doc.Find("div:contains('Jūs vykdote slayer užduotį')").Length() > 0
		if !slayerInProgress {
			err := enableSlayer(p, slayer)
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			return obj.Perform(p, settings)
		}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Pulti')").Attr("href")
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

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	foundElements := doc.Find("div:contains('Taip negalima! turite eiti atgal ir vėl pulti!')").Length()
	if foundElements > 0 {
		return obj.Perform(p, settings)
	}

	// Take some variables
	successWon := doc.Find("div:contains('Nukovėte')").Length() > 0
	successLoss := doc.Find("div:contains('Jūs pralaimėjote')").Length() > 0

	// If not expected - throw error
	if !successWon && !successLoss {
		module.DumpHTML(doc)
		return &module.Result{CanRepeat: false, Error: errors.New("not sure where we are")}
	}

	// Check if we can find health bar and reheal accordingly
	if _, found := settings["eating"]; found {
		threshold, _ := strconv.Atoi(settings["eating_threshold"])
		if threshold == 0 {
			threshold = 50
		}
		// Find progress bar which contains max available health and current health
		val, found := doc.Find("img.hp[src^='graph.php'][src$='c=1']").Eq(1).Attr("src")
		if !found {
			return &module.Result{CanRepeat: false, Error: errors.New("health bar not found")}
		}
		log.Println("src", val)
		val = strings.ReplaceAll(val, "graph.php?", "")
		valPairs := strings.Split(val, "&") // Yes, this is not &amp;
		var remainingHealth, maxHealth int
		for _, v := range valPairs {
			valPairParts := strings.Split(v, "=")
			if valPairParts[0] == "iki" {
				maxHealth, _ = strconv.Atoi(valPairParts[1])
			}
			if valPairParts[0] == "yra" {
				remainingHealth, _ = strconv.Atoi(valPairParts[1])
			}
		}
		if maxHealth == 0 {
			return &module.Result{CanRepeat: false, Error: errors.New("failed to read health bar")}
		}
		log.Println("curr health", int(remainingHealth/maxHealth*100), "threshold", threshold)
		if remainingHealth/maxHealth*100 <= threshold {
			noFoodLeft, err := eating.Eat(p, settings["eating"]) // This function goes on loop, so call this once
			if err != nil {
				return &module.Result{CanRepeat: false, Error: err}
			}
			if successLoss {
				return &module.Result{CanRepeat: false, Error: errors.New("fight lost")}
			}
			if successWon {
				return &module.Result{CanRepeat: !noFoodLeft, Error: nil}
			}
		}
		if successLoss {
			return &module.Result{CanRepeat: false, Error: errors.New("fight lost")}
		}
		if successWon {
			return &module.Result{CanRepeat: true, Error: nil}
		}
	}

	// If action was a success
	if successWon {
		return &module.Result{CanRepeat: true, Error: nil}
	}

	// If action was a success, but fight was lost
	if successLoss {
		return &module.Result{CanRepeat: false, Error: errors.New("fight lost")}
	}

	// If actioned too fast
	foundElements = doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length()
	if foundElements > 0 {
		log.Println("actioned too fast, retrying...")
		return obj.Perform(p, settings)
	}

	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("kovojimas", &Kovojimas{})
}

func enableSlayer(p *player.Player, slayer string) error {
	path := "/slayer.php?{{ creds }}"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return err
	}

	// Check if successfully enabled
	if doc.Find("div:contains('Užduotis atlikta!')").Length() > 0 {
		return enableSlayer(p, slayer)
	}

	path = "/slayer.php?{{ creds }}&id=task&nr=" + slayer

	// Download page that contains unique action link
	doc, err = p.Navigate(path, false)
	if err != nil {
		return err
	}

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	foundElements := doc.Find("div:contains('Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!')").Length()
	if foundElements > 0 {
		return enableSlayer(p, slayer)
	}

	// Check if we've got a reward for previously completed slayer contract
	if doc.Find("div:contains('Jums sėkmingai paskirta užduotis! Grįžę atgal rasite daugiau informacijos apie užduotį.')").Length() >= 0 {
		return nil
	}

	module.DumpHTML(doc)
	return errors.New("slayer logic failed")
}