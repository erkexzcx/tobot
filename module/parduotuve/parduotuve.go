package parduotuve

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"tobot/module"
	"tobot/player"
)

type Parduotuve struct{}

var allowedSettings = map[string][]string{
	"action": {"pirkti", "parduoti"},
}

// Regenerate list: go run cmd/tobot/main.go -shop
var itemsPagesMap = map[string]string{
	"K1":    "1",
	"K2":    "1",
	"K3":    "1",
	"K4":    "1",
	"K5":    "1",
	"K6":    "1",
	"K7":    "1",
	"K8":    "1",
	"K9":    "1",
	"K10":   "1",
	"K11":   "1",
	"K12":   "1",
	"K13":   "1",
	"K14":   "1",
	"K15":   "1",
	"K16":   "1",
	"K17":   "1",
	"K18":   "1",
	"K19":   "1",
	"K20":   "1",
	"K21":   "1",
	"K22":   "1",
	"K23":   "1",
	"K24":   "1",
	"K25":   "1",
	"K26":   "1",
	"K27":   "1",
	"K28":   "1",
	"K29":   "1",
	"K30":   "1",
	"L1":    "12",
	"L2":    "12",
	"L3":    "12",
	"L4":    "12",
	"L5":    "12",
	"L6":    "12",
	"L7":    "12",
	"L8":    "12",
	"L9":    "12",
	"L10":   "12",
	"L11":   "12",
	"ST1":   "13",
	"ST2":   "13",
	"ST3":   "13",
	"ML1":   "32",
	"ML2":   "32",
	"ML3":   "32",
	"ML4":   "32",
	"ML5":   "32",
	"ML6":   "32",
	"ML7":   "32",
	"ML8":   "32",
	"ML9":   "32",
	"ML10":  "32",
	"ML11":  "32",
	"MM1":   "36",
	"S1":    "2",
	"S2":    "2",
	"S3":    "2",
	"S4":    "2",
	"S5":    "2",
	"S6":    "2",
	"S7":    "2",
	"S8":    "2",
	"S9":    "2",
	"S10":   "2",
	"S11":   "2",
	"S12":   "2",
	"S13":   "2",
	"S14":   "2",
	"S15":   "2",
	"D1":    "25",
	"D2":    "25",
	"D3":    "25",
	"D4":    "25",
	"D5":    "25",
	"D6":    "25",
	"D7":    "25",
	"D8":    "25",
	"D9":    "25",
	"D10":   "25",
	"D11":   "25",
	"D12":   "25",
	"D13":   "25",
	"D14":   "25",
	"D15":   "25",
	"W1":    "28",
	"W2":    "28",
	"W3":    "28",
	"W4":    "28",
	"W5":    "28",
	"W6":    "28",
	"W7":    "28",
	"W8":    "28",
	"W9":    "28",
	"W10":   "28",
	"W11":   "28",
	"W12":   "28",
	"W13":   "28",
	"W14":   "28",
	"W15":   "28",
	"H1":    "19",
	"H2":    "19",
	"H3":    "19",
	"H4":    "19",
	"H5":    "19",
	"H6":    "19",
	"H7":    "19",
	"H8":    "19",
	"H9":    "19",
	"H10":   "19",
	"H11":   "19",
	"H12":   "19",
	"H13":   "19",
	"H14":   "19",
	"H15":   "19",
	"PI1":   "22",
	"PI2":   "22",
	"PI3":   "22",
	"PI4":   "22",
	"PI5":   "22",
	"PI6":   "22",
	"PI7":   "22",
	"PI8":   "22",
	"PI9":   "22",
	"PI10":  "22",
	"PI11":  "22",
	"PI12":  "22",
	"PI13":  "22",
	"AP1":   "23",
	"AP2":   "23",
	"AP3":   "23",
	"AP4":   "23",
	"AP5":   "23",
	"AP6":   "23",
	"AP7":   "23",
	"AP8":   "23",
	"AP9":   "23",
	"AP10":  "23",
	"AP11":  "23",
	"AP12":  "23",
	"AP13":  "23",
	"BA1":   "24",
	"BA2":   "24",
	"BA3":   "24",
	"BA4":   "24",
	"BA5":   "24",
	"BA6":   "24",
	"BA7":   "24",
	"BA8":   "24",
	"BA9":   "24",
	"BA10":  "24",
	"BA11":  "24",
	"BA12":  "24",
	"BA13":  "24",
	"HE1":   "33",
	"HE2":   "33",
	"HE3":   "33",
	"HE4":   "33",
	"HE5":   "33",
	"HE6":   "33",
	"HE7":   "33",
	"HE8":   "33",
	"HE9":   "33",
	"HE10":  "33",
	"HE11":  "33",
	"HE12":  "33",
	"HE13":  "33",
	"LI1":   "34",
	"LI2":   "34",
	"LI3":   "34",
	"LI4":   "34",
	"LI5":   "34",
	"LI6":   "34",
	"LI7":   "34",
	"LI8":   "34",
	"LI9":   "34",
	"LI10":  "34",
	"LI11":  "34",
	"LI12":  "34",
	"LI13":  "34",
	"SI1":   "35",
	"SI2":   "35",
	"SI3":   "35",
	"SI4":   "35",
	"SI5":   "35",
	"SI6":   "35",
	"SI7":   "35",
	"SI8":   "35",
	"SI9":   "35",
	"SI10":  "35",
	"SI11":  "35",
	"SI12":  "35",
	"SI13":  "35",
	"Z1":    "7",
	"Z2":    "7",
	"Z3":    "7",
	"Z4":    "7",
	"Z5":    "7",
	"Z6":    "7",
	"Z7":    "7",
	"Z8":    "7",
	"Z9":    "7",
	"Z10":   "7",
	"Z11":   "7",
	"Z12":   "7",
	"Z13":   "7",
	"Z14":   "7",
	"Z15":   "7",
	"Z16":   "7",
	"KZ1":   "8",
	"KZ2":   "8",
	"KZ3":   "8",
	"KZ4":   "8",
	"KZ5":   "8",
	"KZ6":   "8",
	"KZ7":   "8",
	"KZ8":   "8",
	"KZ9":   "8",
	"KZ10":  "8",
	"KZ11":  "8",
	"KZ12":  "8",
	"KZ13":  "8",
	"KZ14":  "8",
	"MS1":   "9",
	"MS2":   "9",
	"MS3":   "9",
	"MS4":   "9",
	"MS5":   "9",
	"MS6":   "9",
	"MS7":   "9",
	"MS8":   "9",
	"MS9":   "9",
	"MS10":  "9",
	"MS11":  "9",
	"MS12":  "9",
	"MS13":  "9",
	"MK1":   "10",
	"MK2":   "10",
	"MK3":   "10",
	"MK4":   "10",
	"MK5":   "10",
	"MK6":   "10",
	"MK7":   "10",
	"MK8":   "10",
	"MK9":   "10",
	"MK10":  "10",
	"MK11":  "10",
	"MK12":  "10",
	"MK13":  "10",
	"KGR1":  "46",
	"KGR2":  "46",
	"KGR3":  "46",
	"KGR4":  "46",
	"KGR5":  "46",
	"KGR6":  "46",
	"KGR7":  "46",
	"KGR8":  "46",
	"KGR9":  "46",
	"KGR10": "46",
	"KGR11": "46",
	"KGR12": "46",
	"MA1":   "11",
	"MA2":   "11",
	"MA3":   "11",
	"MA4":   "11",
	"MA5":   "11",
	"MA6":   "11",
	"MA7":   "11",
	"MA8":   "11",
	"MA9":   "11",
	"MA10":  "11",
	"MA11":  "11",
	"MA12":  "11",
	"MA13":  "11",
	"MA14":  "11",
	"MA15":  "11",
	"MA16":  "11",
	"O1":    "14",
	"O2":    "14",
	"O3":    "14",
	"O4":    "14",
	"O5":    "14",
	"O6":    "14",
	"O7":    "14",
	"O8":    "14",
	"O9":    "14",
	"O10":   "14",
	"O11":   "14",
	"O12":   "14",
	"O13":   "14",
	"O14":   "14",
	"O15":   "14",
	"O16":   "14",
	"B1":    "15",
	"B2":    "15",
	"B3":    "15",
	"B4":    "15",
	"B5":    "15",
	"B6":    "15",
	"B7":    "15",
	"B8":    "15",
	"B9":    "15",
	"B10":   "15",
	"B11":   "15",
	"B12":   "15",
	"B13":   "15",
	"B14":   "15",
	"B15":   "15",
	"B16":   "15",
	"DS1":   "26",
	"DS2":   "26",
	"NM1":   "20",
	"NM2":   "20",
	"NM3":   "20",
	"NM4":   "20",
	"NM5":   "20",
	"NM6":   "20",
	"NM7":   "20",
	"NM8":   "20",
	"NM9":   "20",
	"NM10":  "20",
	"NM11":  "20",
	"NM12":  "20",
	"NM13":  "20",
	"AM1":   "21",
	"AM2":   "21",
	"AM3":   "21",
	"AM4":   "21",
	"AM5":   "21",
	"AM6":   "21",
	"AM7":   "21",
	"AM8":   "21",
	"AM9":   "21",
	"AM10":  "21",
	"AM11":  "21",
	"AM12":  "21",
	"AM13":  "21",
	"M1":    "3",
	"M2":    "3",
	"ME1":   "4",
	"ME2":   "4",
	"ME3":   "4",
	"ME4":   "4",
	"ME5":   "4",
	"P1":    "5",
	"P2":    "5",
	"P3":    "5",
	"P4":    "5",
	"P5":    "5",
	"P6":    "5",
	"P7":    "5",
	"P8":    "5",
	"P9":    "5",
	"P10":   "5",
	"P11":   "5",
	"P12":   "5",
	"P13":   "5",
	"P14":   "5",
	"P15":   "5",
	"P16":   "5",
	"AX1":   "6",
	"AX2":   "6",
	"AX3":   "6",
	"AX4":   "6",
	"AX5":   "6",
	"AX6":   "6",
	"AX7":   "6",
	"AGA1":  "39",
	"AGA2":  "39",
	"AGA3":  "39",
	"AGA4":  "39",
	"AGA5":  "39",
	"AGA6":  "39",
	"KGA1":  "40",
	"KGA2":  "40",
	"KGA3":  "40",
	"KGA4":  "40",
	"KGA5":  "40",
	"KGA6":  "40",
	"KGA7":  "40",
	"KGA8":  "40",
	"KGA9":  "40",
	"KGA10": "40",
	"KGA11": "40",
	"KGA12": "40",
	"KGA13": "40",
	"KGA14": "40",
	"KGA15": "40",
	"KGA16": "40",
	"KOC1":  "41",
	"KOC2":  "41",
	"KOC3":  "41",
	"KOC4":  "41",
	"KOC5":  "41",
	"KOC6":  "41",
	"KOC7":  "41",
	"KOC8":  "41",
	"KOC9":  "41",
	"KOC10": "41",
	"KOC11": "41",
	"NB1":   "17",
	"NB2":   "17",
	"NB3":   "17",
	"NB4":   "17",
	"NB5":   "17",
	"NB6":   "17",
	"NB7":   "17",
	"NB8":   "17",
	"NB9":   "17",
	"NB10":  "17",
	"NB11":  "17",
	"NB12":  "17",
	"NB13":  "17",
	"NB14":  "17",
	"AB1":   "18",
	"AB2":   "18",
	"AB3":   "18",
	"AB4":   "18",
	"AB5":   "18",
	"AB6":   "18",
	"AB7":   "18",
	"AB8":   "18",
	"AB9":   "18",
	"AB10":  "18",
	"AB11":  "18",
	"AB12":  "18",
	"AB13":  "18",
	"AB14":  "18",
	"AL1":   "16",
	"AL2":   "16",
	"AL3":   "16",
	"AL4":   "16",
	"AL5":   "16",
	"AL6":   "16",
	"AL7":   "16",
	"AL8":   "16",
	"AL9":   "16",
	"AL10":  "16",
	"AL11":  "16",
	"AL12":  "16",
	"AL13":  "16",
	"AL14":  "16",
	"ZI1":   "27",
	"ZI2":   "27",
	"ZI3":   "27",
	"ZI4":   "27",
	"ZI5":   "27",
	"ZI6":   "27",
	"ZI7":   "27",
	"ZI8":   "27",
	"ZI9":   "27",
	"ZI10":  "27",
	"ZI11":  "27",
	"ZI12":  "27",
	"ZI13":  "27",
	"ZI14":  "27",
	"SK1":   "29",
	"SK2":   "29",
	"SK3":   "29",
	"SK4":   "29",
	"SK5":   "29",
	"SK6":   "29",
	"SK7":   "29",
	"SK8":   "29",
	"SK9":   "29",
	"SK10":  "29",
	"SK11":  "29",
	"SK12":  "29",
	"SK13":  "29",
	"SK14":  "29",
	"SK15":  "29",
	"SK16":  "29",
	"SK17":  "29",
	"SK18":  "29",
	"SK19":  "29",
	"SK20":  "29",
	"PE1":   "45",
	"PE2":   "45",
	"PE3":   "45",
	"PE4":   "45",
	"PE5":   "45",
	"PE6":   "45",
	"PE7":   "45",
	"PE8":   "45",
	"PE9":   "45",
	"PE10":  "45",
	"PE11":  "45",
	"PE12":  "45",
	"PE13":  "45",
	"PE14":  "45",
	"PE15":  "45",
	"PE16":  "45",
	"SE1":   "30",
	"SE2":   "30",
	"SE3":   "30",
	"SE4":   "30",
	"SE5":   "30",
	"SE6":   "30",
	"SE7":   "30",
	"SE8":   "30",
	"SE9":   "30",
	"SE10":  "30",
	"SE11":  "30",
	"SE12":  "30",
	"SE13":  "30",
	"SE14":  "30",
	"SE15":  "30",
	"SE16":  "30",
	"SE17":  "30",
	"SE18":  "30",
	"SE19":  "30",
	"SE20":  "30",
	"GR1":   "37",
	"GR2":   "37",
	"GR3":   "37",
	"GR4":   "37",
	"GR5":   "37",
	"GR6":   "37",
	"GR7":   "37",
	"GR8":   "37",
	"GR9":   "37",
	"GR10":  "37",
	"GR11":  "37",
	"GR12":  "37",
	"UO1":   "44",
	"UO2":   "44",
	"UO3":   "44",
	"UO4":   "44",
	"UO5":   "44",
	"UO6":   "44",
	"UO7":   "44",
	"UO8":   "44",
	"UO9":   "44",
	"UO10":  "44",
	"UO11":  "44",
	"UO12":  "44",
	"PO1":   "31",
	"PO2":   "31",
	"PO3":   "31",
	"PO4":   "31",
	"PO5":   "31",
	"PO6":   "31",
	"PO7":   "31",
	"PO8":   "31",
	"PO9":   "31",
	"PO10":  "31",
	"PO11":  "31",
	"PO12":  "31",
	"PO13":  "31",
	"PO14":  "31",
	"PO15":  "31",
	"PO16":  "31",
	"PO17":  "31",
	"PO18":  "31",
	"PO19":  "31",
	"PO20":  "31",
	"PA1":   "38",
	"PA2":   "38",
	"PA3":   "38",
	"PA4":   "38",
	"PA5":   "38",
	"PA6":   "38",
	"PA7":   "38",
	"PA8":   "38",
	"PA9":   "38",
	"PA10":  "38",
	"PA11":  "38",
	"PA12":  "38",
	"SPA1":  "43",
	"SPA2":  "43",
	"SPA3":  "43",
	"SPA4":  "43",
	"SPA5":  "43",
	"SPA6":  "43",
	"SPA7":  "43",
	"SPA8":  "43",
	"SPA9":  "43",
	"SPA10": "43",
	"SPA11": "43",
	"SPA12": "43",
	"MANA1": "42",
	"MANA2": "42",
	"MANA3": "42",
	"MANA4": "42",
	"MANA5": "42",
	"MANA6": "42",
	"MANA7": "42",
}

func (obj *Parduotuve) Validate(settings map[string]string) error {
	// Check for missing keys
	for k := range allowedSettings {
		_, found := settings[k]
		if !found {
			return errors.New("missing key '" + k + "'")
		}
	}

	for k, v := range settings {
		if strings.HasPrefix(k, "_") || k == "item" || k == "amount" {
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

	// Check 'item'
	item, found := settings["item"]
	if !found {
		return errors.New("missing key 'item'")
	}
	_, found = itemsPagesMap[item]
	if !found {
		return errors.New("unrecognized value of key 'item'")
	}

	//Check 'amount'
	amount, found := settings["amount"]
	if !found {
		return errors.New("missing key 'amount'")
	}
	_, err := strconv.Atoi(amount)
	if err != nil {
		return errors.New("unrecognized value of key 'amount'")
	}

	return nil
}

func (obj *Parduotuve) Perform(p *player.Player, settings map[string]string) *module.Result {
	if settings["action"] == "pirkti" {
		return buy(p, settings)
	}
	return sell(p, settings)
}

var regexPirktiMax = regexp.MustCompile(`Daugiausia galite nusipirkti šių daiktų: <b>(\d+)</b>`)

func buy(p *player.Player, settings map[string]string) *module.Result {
	page := itemsPagesMap[settings["item"]]
	amount, _ := strconv.Atoi(settings["amount"])

	path := "/parda.php?{{ creds }}&id=pirkt&ka=" + settings["item"] + "&page=" + page
	pathSubmit := "/parda.php?{{ creds }}&id=perku&ka=" + settings["item"] + "&page=" + page

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find how many we can buy
	code, err := doc.Html()
	if err != nil {
		return buy(p, settings) // retry
	}
	maxToBuyMatch := regexPirktiMax.FindStringSubmatch(code)
	if len(maxToBuyMatch) != 2 {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to find count of available to buy items")}
	}
	maxToBuy := maxToBuyMatch[1]

	// Convert string number to actual int
	maxToBuyInt, err := strconv.Atoi(maxToBuy)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to understand max number of items available to buy")}
	}

	buyAmount := maxToBuyInt
	if amount <= 0 {
		buyAmount = maxToBuyInt + amount // Adding negative number becomes subtraction
	}
	if buyAmount < 1 {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to buy " + fmt.Sprint(buyAmount) + " amount of items")}
	}

	params := url.Values{}
	params.Add("kiekis", fmt.Sprint(buyAmount))
	params.Add("null", "Pirkti")
	body := strings.NewReader(params.Encode())

	// Submit request
	doc, err = p.Submit(pathSubmit, body)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If action was a success
	foundElements := doc.Find("div:contains('Daiktai nupirkti, išleidote ')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	html, _ := doc.Html()
	log.Println(html)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func sell(p *player.Player, settings map[string]string) *module.Result {
	page := itemsPagesMap[settings["item"]]
	amount, _ := strconv.Atoi(settings["amount"])

	path := "/parda.php?{{ creds }}&id=parduot&ka=" + settings["item"] + "&page=" + page
	pathSubmit := "/parda.php?{{ creds }}&id=parduodu&ka=" + settings["item"] + "&page=" + page

	// Download page that contains unique action link
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Find how many we can sell
	maxToSell, found := doc.Find("form > input[name='kiekis'][type='hidden']").Attr("value")
	if !found {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to find count of available to sell items")}
	}

	// Convert string number to actual int
	maxToSellInt, err := strconv.Atoi(maxToSell)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to understand max number of items available to sell")}
	}

	sellAmount := maxToSellInt
	if amount <= 0 {
		sellAmount = maxToSellInt + amount // Adding negative number becomes subtraction
	}
	if sellAmount < 0 {
		return &module.Result{CanRepeat: false, Error: errors.New("unable to sell " + fmt.Sprint(sellAmount) + " amount of items")}
	}
	if sellAmount == 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	params := url.Values{}
	params.Add("kiekis", fmt.Sprint(sellAmount))
	params.Add("null", "Parduoti")
	body := strings.NewReader(params.Encode())

	// Submit request
	doc, err = p.Submit(pathSubmit, body)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If action was a success
	foundElements := doc.Find("div:contains('Daiktai parduoti, gavote ')").Length()
	if foundElements > 0 {
		return &module.Result{CanRepeat: false, Error: nil}
	}

	html, _ := doc.Html()
	log.Println(html)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func init() {
	module.Add("parduotuve", &Parduotuve{})
}
