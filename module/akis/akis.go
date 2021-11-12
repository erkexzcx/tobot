package akis

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"tobot/module"
	"tobot/player"
)

type Akis struct{}

func (obj *Akis) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

func (obj *Akis) Perform(p *player.Player, settings map[string]string) *module.Result {
	currentCoins := 10
	pathSubmit := "/kazino.php?{{ creds }}&id=akis2"

	for {
		// Antiflood protection
		time.Sleep(1500 * time.Millisecond)

		params := url.Values{}
		params.Add("nr", fmt.Sprint(currentCoins))
		params.Add("null", "Losti")
		body := strings.NewReader(params.Encode())

		// Submit request
		log.Printf("betting %d coins\n", currentCoins)
		doc, err := p.Submit(pathSubmit, body)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}

		// Check if noone wins
		if doc.Find("div:contains('Lygiosios, niekas nieko nelaimejo ir nieko nepralaimejo.')").Length() > 0 {
			log.Println("noone wins, retrying")
			continue
		}

		// Check if antiflood
		if doc.Find("div:contains('Antiflood! Bandykite vel po keleto sekundziu..')").Length() > 0 {
			log.Println("antiflood, retrying")
			continue
		}

		// Check if lost
		if doc.Find("div:contains('Deja, taciau jus pralaimejote')").Length() > 0 {
			log.Println("lost, retrying with double coins")
			currentCoins += currentCoins
			continue
		}

		// Check if not enough coins
		if doc.Find("div:contains('Tiek zetonu neturi!')").Length() > 0 {
			log.Println("not enough coins")
			return &module.Result{CanRepeat: false, Error: errors.New("not enough coins")}
		}

		// Check if won
		if doc.Find("div:contains('Sveikiname! Jus laimejote')").Length() > 0 {
			log.Println("won!")
			return &module.Result{CanRepeat: false, Error: err}
		}

		module.DumpHTML(doc)
		return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
	}
}

func init() {
	module.Add("akis", &Akis{})
}
