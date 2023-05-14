package player

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"tobot/comms"
	"tobot/config"

	"github.com/PuerkitoBio/goquery"
)

// Defines minimum amount of time needed to wait to not get "clicked too fast" error
const MIN_WAIT_TIME = 625 * time.Millisecond

// Navigate is used to navigate & perform activities in-game.
func (p *Player) Navigate(path string, action bool) (goqueryDocument *goquery.Document, wrongDoc bool, err error) {
	return p.openLink(path, action, "GET", nil)
}

// Submit is used to submit forms in-game.
func (p *Player) Submit(path string, body io.Reader) (goqueryDocument *goquery.Document, wrongDoc bool, err error) {
	return p.openLink(path, false, "POST", body)
}

func (p *Player) openLink(path string, action bool, method string, body io.Reader) (goqueryDocument *goquery.Document, wrongDoc bool, err error) {
	// Remember the time of this as soon as possible
	timeNow := time.Now()

	bodyPassed := true
	if body == nil {
		bodyPassed = false
	}
	p.Log.Debugf("Performing request: {path: %s, action: %t, method: %s, body passed: %t}\n", path, action, method, bodyPassed)

	// Mandatory wait before opening any link in the game
	var timeToSleep time.Duration
	if action {
		timeToSleep = p.timeUntilAction.Sub(timeNow)
	} else {
		timeToSleep = p.timeUntilNavigation.Sub(timeNow)
	}
	p.Log.Debugf("Sleeping for %s before performing request\n", timeToSleep)
	time.Sleep(timeToSleep)

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
		p.Log.Warningf("Failed to perform HTTP request (sleeping for 5 seconds and re-trying request): %s\n", err.Error())
		time.Sleep(5 * time.Second)
		return nil, true, err // We don't know if server received the request - let's assume we need to re-try
	}
	defer resp.Body.Close()

	// Mark timestamp when doc was downloaded
	timeNow = time.Now()

	// Create GoQuery document out of response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		p.Log.Warningf("Failed to get GoQuery document from response body (sleeping for 5 seconds and re-trying request): %s\n", err.Error())
		time.Sleep(5 * time.Second)
		return doc, true, nil // Failure downloading response body - server definitely processed the request
	}

	// Remember until when we have to wait before opening another link
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - *p.Config.Settings.MinRTT)
	if action {
		p.timeUntilAction = timeNow.Add(p.extractActionWaitTime(doc) - *p.Config.Settings.MinRTT)
	}
	if p.timeUntilAction.Before(p.timeUntilNavigation) {
		p.timeUntilAction = p.timeUntilNavigation
	}
	p.Log.Debugf("Extracted wait times: {navigation: %s, action: %s}\n", p.timeUntilNavigation.Sub(timeNow), p.timeUntilAction.Sub(timeNow))

	// Check if we clicked too fast
	if doc.Find("b:contains('NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!')").Length() > 0 {
		p.Log.Warning("Received 'Clicked too fast' error (sleeping for 15 seconds and re-trying request)")
		time.Sleep(time.Minute)
		return p.openLink(path, action, method, body)
	}

	// Checks if there are new PMs
	if doc.Find("a[href*='id=pm']:contains('Yra naujų PM')").Length() > 0 {
		p.Log.Debug("New PM detected!")
		if err := p.dealWithPMs(); err != nil {
			p.Log.Warningf("Failed to manage new PMs: %s\n", err.Error())
			comms.SendMessageToTelegram(fmt.Sprintf("Failed to manage new PMs: %s", err.Error()))
		}
	}

	// Check if anti-cheat check is present
	if doc.Find("div:contains('Paspauskite žemiau esančią šią spalvą:')").Length() > 0 {
		p.Log.Debug("Anti-cheat check detected!")
		err := p.solveAnticheat(doc)
		if err == nil {
			p.Log.Info("Successfully solved anti-cheat check")
		} else {
			p.Log.Warningf("Failed to solve anti-cheat check: %s\n", err.Error())
		}
		// The problem with this anti-cheat check is that your requested action (e.g. POST request) is performed,
		// such as buying X amount of X items, but due to anti-cheat page, we cannot know if action was a success
		// or a fail. This should be handled by the caller of this function.
		return doc, true, nil
	}

	// Check if earlier anti-cheat failed
	if doc.Find("div:contains('Praėjo spalvos paspaudimo laikas')").Length() > 0 {
		p.Log.Warning("Anti-cheat check timeout detected! (re-trying request)")
		return doc, true, nil // TODO - I assume that it's the same behavior as anti-cheat check window?
	}

	// Check if player is banned
	if doc.Find("div:contains('Jūs užbanintas.')").Length() > 0 {
		return nil, false, errors.New("player banned")
	}

	// Check if misconfiguration/marked as bot
	if doc.Find("div:contains('Sistema nustatė, jog jūs jungiates per kitą serverį, todėl greičiausiai bandote naudotis autokėlėju.')").Length() > 0 {
		return nil, false, errors.New("misconfiguration or your IP/configuration is marked as bot")
	}

	// Check if bad credentials or player does not exist (deleted)
	// This check must be after all other checks
	if doc.Find("div:contains('Blogi duomenys!')").Length() > 0 {
		registrationMux.Lock()
		defer registrationMux.Unlock()
		if config.CreatePlayers {
			return nil, false, errors.New("invalid credentials or player does not exist (deleted?)")
		}
		err = p.registerPlayer()
		if err != nil {
			time.Sleep(5 * time.Second) // Wait 5 seconds between retries, so we don't DOS server
			return p.openLink(path, action, method, body)
		}
		return nil, true, nil
	}

	return doc, false, nil
}

func (p *Player) extractActionWaitTime(doc *goquery.Document) time.Duration {
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
		p.Log.Debug("Generating sleep durations (become offline) for upcoming sleep")

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
	p.Log.Infof("Sleeping (become offline) for %s", sleepDuration.String())
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
	if randSeeded.Float64() >= probability {
		return time.Duration(0)
	}
	return randomDuration(minDuration, maxDuration)
}

// randomDuration takes two time.Duration values and returns a random duration between them.
func randomDuration(minDuration, maxDuration time.Duration) time.Duration {
	durationDiff := maxDuration - minDuration
	randomFloat := randSeeded.Float64()
	randomDuration := minDuration + time.Duration(randomFloat*float64(durationDiff))
	return randomDuration
}
