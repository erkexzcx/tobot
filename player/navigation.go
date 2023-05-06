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

// Defines minimum amount of time needed to wait to not get "clicked too fast" error
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
	// Mandatory wait before opening any link in the game
	timeNow := time.Now()
	if action {
		time.Sleep(p.timeUntilAction.Sub(timeNow))
	} else {
		time.Sleep(p.timeUntilNavigation.Sub(timeNow))
	}

	// Check if we have to become offline
	if !action && method == "GET" {
		p.manageBecomeOffline()
	}

	// Check if we have to additionally wait before action
	if action && method == "GET" {
		p.manageRandomWait()
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

	// Check if bad credentials or player does not exist (deleted)
	if doc.Find("div:contains('Blogi duomenys!')").Length() > 0 {
		return nil, errors.New("invalid credentials or player does not exist (deleted?)")
	}

	// Check if player is banned
	if doc.Find("div:contains('Jūs užbanintas.')").Length() > 0 {
		return nil, errors.New("player banned")
	}

	// Check if misconfiguration/marked as bot
	if doc.Find("div:contains('Sistema nustatė, jog jūs jungiates per kitą serverį, todėl greičiausiai bandote naudotis autokėlėju.')").Length() > 0 {
		return nil, errors.New("misconfiguration or your IP/configuration is marked as bot")
	}

	// Check if we clicked too fast
	if doc.Find("b:contains('NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!')").Length() > 0 {
		log.Println("[" + p.Config.Nick + "] Clicked too fast and now sleeps for 15 seconds...")
		time.Sleep(15 * time.Second)
		return p.openLink(path, action, method, body)
	}

	// Check if anti-cheat check is present
	if doc.Find("div:contains('Paspauskite žemiau esančią šią spalvą:')").Length() > 0 {
		err := p.solveAnticheat(doc)
		if err != nil {
			log.Println("Successfully solved anti-cheat check")
		} else {
			log.Printf("Failed to solve anti-cheat check: %s", err.Error())
		}
		return p.openLink(path, action, method, body)
	}

	// Checks if there are new PMs
	if doc.Find("a[href*='id=pm']:contains('Yra naujų PM')").Length() > 0 {
		p.dealWithPMs()
	}

	return doc, nil
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

// This function provides ability to constantly go offline (sleep random durations at random intervals)
func (p *Player) manageBecomeOffline() {
	// If not enabled - return
	if !*p.Config.Settings.BecomeOffline.Enabled {
		return
	}

	// Get current timestamp
	timeNow := time.Now()

	// If before sleep period - return
	if timeNow.Before(p.sleepFrom) {
		return
	}

	// If after sleep period - generate new sleep period
	if timeNow.After(p.sleepTo) {
		// Generate new random sleep duration
		sleepDuration := randomDuration(
			(*p.Config.Settings.BecomeOffline.For)[0],
			(*p.Config.Settings.BecomeOffline.For)[1],
		)

		// Generate random duration until sleep should occur
		sleepIn := randomDuration(
			(*p.Config.Settings.BecomeOffline.Every)[0],
			(*p.Config.Settings.BecomeOffline.Every)[1],
		)

		// Update player variables
		p.sleepFrom = time.Now().Add(time.Duration(sleepIn))
		p.sleepTo = p.sleepFrom.Add(time.Duration(sleepDuration))

		return
	}

	// If within sleep period - sleep (become offline)
	sleepDuration := p.sleepTo.Sub(timeNow)
	log.Println("Sleeping for", sleepDuration.String())
	time.Sleep(sleepDuration)
}

// This function allows to add custom additional duration before opening an action link
func (p *Player) manageRandomWait() {
	// If not enabled - return
	if !*p.Config.Settings.RandomizeWait.Enabled {
		return
	}

	// Get random additional duration to wait
	timeToWait := randomDurationWithProbability(
		(*p.Config.Settings.RandomizeWait.WaitVal)[0],
		(*p.Config.Settings.RandomizeWait.WaitVal)[1],
		*p.Config.Settings.RandomizeWait.Probability,
	)

	// Sleep
	time.Sleep(timeToWait)
}

// randomDurationWithProbability takes two time.Duration values, a success rate probability,
// and returns a random duration between them or 0 based on the success rate.
func randomDurationWithProbability(minDuration, maxDuration time.Duration, probability float64) time.Duration {
	if rand.Float64() >= probability {
		return time.Duration(0)
	}
	return randomDuration(minDuration, maxDuration)
}

// randomDuration takes two time.Duration values and returns a random duration between them.
func randomDuration(minDuration, maxDuration time.Duration) time.Duration {
	durationDiff := maxDuration - minDuration
	randomFloat := rand.Float64()
	randomDuration := minDuration + time.Duration(randomFloat*float64(durationDiff))
	return randomDuration
}
