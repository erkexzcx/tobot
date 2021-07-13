package player

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MIN_WAIT_TIME = 555 * time.Millisecond

// Navigate is used to navigate & perform activities in-game. It cannot click too fast, tracks new PMs
func (p *Player) Navigate(path string, action bool) (*goquery.Document, error) {
	p.manageBecomeOffline()

	// Wait until performing HTTP request
	timeNow := time.Now()
	p.timeUntilMux.Lock()
	waitUntil := p.timeUntilNavigation
	if action {
		waitUntil = p.timeUntilAction
	}
	p.timeUntilMux.Unlock()
	timeToWait := waitUntil.Sub(timeNow)
	if timeToWait < MIN_WAIT_TIME-p.minRTTTime {
		timeToWait = MIN_WAIT_TIME - p.minRTTTime
	}
	time.Sleep(timeToWait)

	// Perform HTTP request and get response
	resp, err := p.httpRequest("GET", p.fullLink(path), nil)
	if err != nil {
		log.Println("Failure occurred (#1): " + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.Navigate(path, action)
	}
	defer resp.Body.Close()

	// Mark timestamp when doc was downloaded
	timeNow = time.Now()

	// Create Goquery document out of response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Failure occurred (#2): " + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.Navigate(path, action)
	}

	// Mark wait time
	p.timeUntilMux.Lock()
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - p.minRTTTime)
	if action {
		p.timeUntilAction = timeNow.Add(p.extractWaitTime(doc))
	}
	p.timeUntilMux.Unlock()

	// Try again if clicked too fast!
	if isTooFast(doc) {
		r := getRandomInt(3123, 8765)
		log.Println("Clicked too fast! Sleeping for " + fmt.Sprintf("%.2f", float64(r)/1000) + "s and trying again...")
		time.Sleep(time.Duration(r) * time.Millisecond)
		return p.Navigate(path, action)
	}

	// Check if landed in anti-cheat check page
	if isAnticheatPage(doc) {
		res := p.solveAnticheat(doc)
		if !res {
			log.Println("Anti cheat procedure failed...")
		}
		return p.Navigate(path, action)
	}

	// Check if banned
	if isBanned(doc) {
		p.NotifyTelegram("player banned")
		return nil, errors.New("player banned")
	}

	// Check if has new PMs - for now notify the user and panic
	if hasNewPM(doc) {
		m, err := p.getLastPM()
		if err != nil {
			panic(err)
		}

		// Ignore @sistema
		if m.from == "" {
			return p.Navigate(path, action)
		}

		// Ignore non-moderators
		if !m.moderator {
			time.Sleep(5 * time.Second)
			return p.Navigate(path, action)
		}

		// Attempt to autoreply
		replyMsg, ignore := generateReply(m.text)

		if ignore {
			time.Sleep(8 * time.Second)
			return p.Navigate(path, action)
		}

		if replyMsg != "" {
			time.Sleep(15 * time.Second)
			p.sendPM(m.from, replyMsg)
			return p.Navigate(path, action)
		}

		if m.moderator {
			p.NotifyTelegram("User '*" + m.from + "' says: " + m.text)
		} else {
			p.NotifyTelegram("User '" + m.from + "' says: " + m.text)
		}

		// Lock - Telegram will unlock it
		p.waitingPMMux.Lock()
		p.waitingPM = true
		p.waitingPMMux.Unlock()

		// Wait
		for {
			p.waitingPMMux.Lock()
			waitingState := p.waitingPM
			p.waitingPMMux.Unlock()

			if !waitingState {
				break
			}

			time.Sleep(300)
		}

		return p.Navigate(path, action)
	}

	return doc, nil
}

// Submit is used to submit forms in-game.
func (p *Player) Submit(path string, body io.Reader) (*goquery.Document, error) {
	p.manageBecomeOffline()

	fullLink := p.fullLink(path)

	// Wait until performing HTTP request
	p.timeUntilMux.Lock()
	timeToWait := time.Until(p.timeUntilNavigation)
	p.timeUntilMux.Unlock()
	time.Sleep(timeToWait)

	// Perform HTTP request and get response
	resp, err := p.httpRequest("POST", fullLink, body)
	if err != nil {
		log.Println("Failure occurred (#1): " + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.Submit(path, body)
	}
	defer resp.Body.Close()

	// Mark timestamp when doc was downloaded
	timeNow := time.Now()

	// Create Goquery document out of response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Failure occurred (#2): " + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.Submit(path, body)
	}

	// Try again if clicked too fast!
	if isTooFast(doc) {
		r := getRandomInt(3123, 8765)
		log.Println("Clicked too fast! Sleeping for " + fmt.Sprintf("%.2f", float64(r)/1000) + "s and trying again...")
		time.Sleep(time.Duration(r) * time.Millisecond)
		return p.Submit(path, body)
	}

	// Mark wait time
	p.timeUntilMux.Lock()
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - p.minRTTTime)
	p.timeUntilMux.Unlock()

	// Check if landed in anti-cheat check page
	if isAnticheatPage(doc) {
		res := p.solveAnticheat(doc)
		if !res {
			log.Println("Anti cheat procedure failed...")
		}
		return p.Submit(path, body)
	}

	// Check if banned
	if isBanned(doc) {
		p.NotifyTelegram("player banned")
		return nil, errors.New("player banned")
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
	// <b>Jūs užbanintas.<br/>
	return doc.Find("div:contains('Sistema nustatė, jog jūs jungiates per kitą serverį, todėl greičiausiai bandote naudotis autokėlėju.')").Length() > 0 ||
		doc.Find("div:contains('Jūs užbanintas.')").Length() > 0
}

func isAnticheatPage(doc *goquery.Document) bool {
	return doc.Find("div:contains('Paspauskite žemiau esančią šią spalvą:')").Length() > 0
}

func (p *Player) extractWaitTime(doc *goquery.Document) time.Duration {
	timeLeft, found := doc.Find("#countdown").Attr("title")
	if !found {
		return MIN_WAIT_TIME - p.minRTTTime
	}
	parsedDuration, err := time.ParseDuration(timeLeft + "s")
	if err != nil {
		panic(err)
	}
	if parsedDuration > MIN_WAIT_TIME-p.minRTTTime {
		return parsedDuration - p.minRTTTime
	}
	return MIN_WAIT_TIME - p.minRTTTime
}

func (p *Player) fullLink(path string) string {
	return p.rootLink + strings.ReplaceAll(path, "{{ creds }}", "nick="+p.nick+"&pass="+p.pass)
}

func getRandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func getRandomInt64(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

func (p *Player) manageBecomeOffline() {
	if !p.becomeOffline {
		return
	}

	if getPausedState() {
		log.Println("Bot stopped by telegram bot...")
		for {
			if !getPausedState() {
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
		log.Println("Bot resumed by telegram bot...")
	}

	timeNow := time.Now()

	if timeNow.After(p.sleepTo) {
		p.updateBecomeOfflineTimes()
		return
	}

	if timeNow.After(p.sleepFrom) && timeNow.Before(p.sleepTo) {
		sleepDuration := p.sleepTo.Sub(timeNow)
		log.Println("Sleeping for", sleepDuration.String())
		time.Sleep(sleepDuration)
		p.updateBecomeOfflineTimes()
		return
	}
}

func (p *Player) updateBecomeOfflineTimes() {
	sleepDuration := getRandomInt64(int64(p.becomeOfflineForFrom), int64(p.becomeOfflineForTo))
	sleepIn := getRandomInt64(int64(p.becomeOfflineEveryFrom), int64(p.becomeOfflineEveryTo))
	p.sleepFrom = time.Now().Add(time.Duration(sleepIn))
	p.sleepTo = p.sleepFrom.Add(time.Duration(sleepDuration))
}
