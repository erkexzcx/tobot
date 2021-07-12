package slayer

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Slayer struct{}

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
	// TODO

	// Sniegynas
	// TODO

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
}

var allowedSlayers = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"5": true,
}

func (obj *Slayer) Validate(settings map[string]string) error {
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
			return errors.New("invalid 'slayer' value")
		}
	}

	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if k != "vs" && k != "slayer" {
			return errors.New("unknown '" + k + "' field")
		}
	}

	return nil
}

var reMatchProgress = regexp.MustCompile(`Atlikta: (\d+) \/ (\d+)`)

func (obj *Slayer) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kova.php?{{ creds }}&id=kova0&vs=" + settings["vs"] + "&psl=" + enemyToPageMap[settings["vs"]]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find action link
	actionLink, found := doc.Find("a[href*='&kd=']:contains('Pulti')").Attr("href")
	if !found {
		module.DumpHTML(doc)
		return &module.Result{CanRepeat: false, Error: errors.New("action button not found")}
	}

	// Find slayer info, perform it and if neeeded - restart this function
	slayerInProgress := doc.Find("div:contains('Jūs vykdote slayer užduotį')").Length() > 0
	_, found = settings["slayer"]
	if !slayerInProgress && found {
		err := enableSlayer(p, settings)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		return obj.Perform(p, settings)
	}
	matches := reMatchProgress.FindStringSubmatch(doc.Text())
	if len(matches) != 3 {
		return &module.Result{CanRepeat: false, Error: errors.New("invalid regex (FIXME)")}
	}
	if matches[1] == matches[2] {
		err := takeReward(p, settings)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		return &module.Result{CanRepeat: false, Error: nil} // stop
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
	foundElements = doc.Find("div:contains('Nukovėte')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: true, Error: nil}
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

func enableSlayer(p *player.Player, settings map[string]string) error {
	path := "/slayer.php?{{ creds }}&id=task&nr=" + settings["slayer"]

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return err
	}

	// Above function might retry in some cases, so if page asks us to go back and try again - lets do it:
	foundElements := doc.Find("div:contains('Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!')").Length()
	if foundElements > 0 {
		return enableSlayer(p, settings)
	}

	// Check if successfully enabled
	foundElement := doc.Find("div:contains('Jums sėkmingai paskirta užduotis! Grįžę atgal rasite daugiau informacijos apie užduotį.')").Length()
	if foundElement == 0 {
		module.DumpHTML(doc)
		return errors.New("failed to enable Slayer contract")
	}

	return nil
}

func takeReward(p *player.Player, settings map[string]string) error {
	path := "/slayer.php?{{ creds }}"

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return err
	}

	// Check if successfully enabled
	foundElement := doc.Find("div:contains('Užduotis atlikta!')").Length()
	if foundElement == 0 {
		module.DumpHTML(doc)
		return errors.New("did not take Slayer reward successfully")
	}

	return nil
}

func init() {
	module.Add("slayer", &Slayer{})
}
