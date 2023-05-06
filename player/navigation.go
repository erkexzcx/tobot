package player

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MIN_WAIT_TIME = 625 * time.Millisecond

// Navigate is used to navigate & perform activities in-game.
func (p *Player) Navigate(path string, action bool) (*goquery.Document, error) {
	return p.openLink(path, action, "GET", nil)
}

// Submit is used to submit forms in-game.
func (p *Player) Submit(path string, body io.Reader) (*goquery.Document, error) {
	return p.openLink(path, false, "POST", body)
}

func (p *Player) openLink(path string, action bool, method string, body io.Reader) (*goquery.Document, error) {
	// Remember the timestamp
	timeNow := time.Now()

	// Check if we have to become offline
	if !action && method == "GET" {
		p.manageBecomeOffline()
	}

	// Wait until performing HTTP request
	if action {
		time.Sleep(p.timeUntilAction.Sub(timeNow))
	} else {
		time.Sleep(p.timeUntilNavigation.Sub(timeNow))
		p.randomWait()
	}

	// Perform HTTP request and get response
	fullLink := p.renderFullLink(path)
	resp, err := p.httpRequest(method, fullLink, body)
	if err != nil {
		log.Println("Failed to perform HTTP request:" + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.openLink(path, action, method, body)
	}
	defer resp.Body.Close()

	// Mark timestamp when doc was downloaded
	timeNow = time.Now()

	// Create GoQuery document out of response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Failed to download HTTP response for GoQuery document:" + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.openLink(path, action, method, body)
	}

	// Remember until when we have to wait before opening another link
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - *p.Config.Settings.MinRTT)
	if action {
		p.timeUntilAction = timeNow.Add(p.extractWaitTime(doc) - *p.Config.Settings.MinRTT)
	} else {
		p.timeUntilAction = p.timeUntilNavigation
	}

	// Checks where did we land
	if isPlayerNotExist(doc) {
		return nil, errors.New("player deleted or does not exist")
	}
	if isBanned(doc) {
		return nil, errors.New("player banned")
	}
	if isTooFast(doc) {
		log.Println("[" + p.Config.Nick + "] Clicked too fast and now sleeps for 15 seconds...")
		time.Sleep(15 * time.Second)
		return p.openLink(path, action, method, body)
	}
	if isAnticheatPage(doc) {
		err := p.solveAnticheat(doc)
		if err != nil {
			log.Println("Successfully solved anti-cheat check")
		} else {
			log.Printf("Failed to solve anti-cheat check: %s", err.Error())
		}
		return p.openLink(path, action, method, body)
	}

	// Checks if there are new PMs
	if hasNewPM(doc) {
		p.dealWithPMs()
	}

	return doc, nil
}

func hasNewPM(doc *goquery.Document) bool {
	return doc.Find("a[href*='id=pm']:contains('Yra naujų PM')").Length() > 0
}

func isTooFast(doc *goquery.Document) bool {
	return doc.Find("b:contains('NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!')").Length() > 0
}

func isBanned(doc *goquery.Document) bool {
	return doc.Find("div:contains('Sistema nustatė, jog jūs jungiates per kitą serverį, todėl greičiausiai bandote naudotis autokėlėju.')").Length() > 0 ||
		doc.Find("div:contains('Jūs užbanintas.')").Length() > 0
}

func isPlayerNotExist(doc *goquery.Document) bool {
	return doc.Find("div:contains('Blogi duomenys!')").Length() > 0
}

func isAnticheatPage(doc *goquery.Document) bool {
	return doc.Find("div:contains('Paspauskite žemiau esančią šią spalvą:')").Length() > 0
}

func (p *Player) extractWaitTime(doc *goquery.Document) time.Duration {
	timeLeft, found := doc.Find("#countdown").Attr("title")
	if !found {
		return MIN_WAIT_TIME
	}
	parsedDuration, err := time.ParseDuration(timeLeft + "s")
	if err != nil {
		panic(err)
	}
	if parsedDuration == 0 {
		return MIN_WAIT_TIME
	}
	return parsedDuration
}

func (p *Player) renderFullLink(path string) string {
	return *p.Config.Settings.RootAddress + strings.ReplaceAll(path, "{{ creds }}", "nick="+p.Config.Nick+"&pass="+p.Config.Pass)
}

func getRandomInt(min, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return random.Intn(max-min) + min
}

func getRandomInt64(min, max int64) int64 {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return random.Int63n(max-min) + min
}

func getRandomForAdditionalWait(min, max int64) int64 {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	number := random.Int63n(max*2-min) + min
	// Below logic helps to be more human like :)
	if number > max {
		return number / 2
	}
	return min
}

func (p *Player) manageBecomeOffline() {
	if p.becomeOfflineEveryTo == 0 && p.becomeOfflineForTo == 0 {
		return
	}

	timeNow := time.Now()

	// If we are past sleep period, generate new period
	if timeNow.After(p.sleepTo) {
		p.updateBecomeOfflineTimes()
		return
	}

	// If we have to sleep - sleep now
	if timeNow.After(p.sleepFrom) && timeNow.Before(p.sleepTo) {
		sleepDuration := p.sleepTo.Sub(timeNow)
		p.Println("Sleeping for", sleepDuration.String())
		time.Sleep(sleepDuration)
		return
	}
}

func (p *Player) updateBecomeOfflineTimes() {
	sleepDuration := getRandomInt64(int64(p.becomeOfflineForFrom), int64(p.becomeOfflineForTo))
	sleepIn := getRandomInt64(int64(p.becomeOfflineEveryFrom), int64(p.becomeOfflineEveryTo))
	p.sleepFrom = time.Now().Add(time.Duration(sleepIn))
	p.sleepTo = p.sleepFrom.Add(time.Duration(sleepDuration))
}

func (p *Player) randomWait() {
	if p.randomizeWaitTo != 0 {
		timeToWait := time.Duration(getRandomForAdditionalWait(int64(p.randomizeWaitFrom), int64(p.randomizeWaitTo)))
		time.Sleep(timeToWait)
	}
}
