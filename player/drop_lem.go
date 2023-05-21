package player

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Find out how much we have (GET request)
// http://tob.lt/zaisti.php?{{ creds }}&id=drop&ka=LEM

// Find out how much we have (POST request)
// http://tob.lt/zaisti.php?{{ creds }}&id=drop2&ka=LEM

func (p *Player) dropAllLEM() error {
	path := "/zaisti.php?{{ creds }}&id=drop&ka=LEM"
	doc, wrongDoc, err := p.Navigate(path, false)
	if err != nil {
		return err
	}
	if wrongDoc {
		p.Log.Warning("Got wrong document during LEM drop, trying again")
		return p.dropAllLEM()
	}

	maxToDrop, found := doc.Find("form > input[name='kiekis'][type='hidden']").Attr("value")
	if !found {
		p.Log.Warning("No LEM drop, exiting")
		return nil // Probably empty page, nothing to drop
	}

	// Build request body
	params := url.Values{}
	params.Add("kiekis", maxToDrop)
	params.Add("null", "Išmesti visus")
	body := strings.NewReader(params.Encode())

	// Submit request
	path = "/zaisti.php?{{ creds }}&id=drop2&ka=LEM"
	doc, wrongDoc, err = p.Submit(path, body)
	if err != nil {
		return err
	}
	if wrongDoc {
		p.Log.Warning("Got wrong document during LEM drop submit, trying again")
		return p.dropAllLEM()
	}

	if doc.Find("div:contains('Daiktai išmesti')").Length() > 0 {
		return nil
	}

	if doc.Find("div:contains('Antiflood! Bandykite už kelių sekundžių.')").Length() > 0 {
		p.Log.Warning("Got too fast click during LEM drop submit, trying again in 5 seconds")
		time.Sleep(5 * time.Second)
		return p.dropAllLEM()
	}

	html, _ := doc.Html()
	fmt.Println(html)

	return errors.New("unknown error occurred during dropping LEMs")
}
