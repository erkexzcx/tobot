package player

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"
	"tobot/telegram"

	"github.com/PuerkitoBio/goquery"
)

const MIN_WAIT_TIME = 625 * time.Millisecond

// Navigate is used to navigate & perform activities in-game. It cannot click too fast, tracks new PMs
func (p *Player) Navigate(path string, action bool) (*goquery.Document, error) {
	// Mark current time
	timeNow := time.Now()

	p.manageBecomeOffline()

	// Wait until performing HTTP request
	if action {
		time.Sleep(p.timeUntilAction.Sub(timeNow))
	} else {
		time.Sleep(p.timeUntilNavigation.Sub(timeNow))
		p.randomWait()
	}

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
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - p.minRTT)
	if action {
		p.timeUntilAction = timeNow.Add(p.extractWaitTime(doc) - p.minRTT)
	}
	if p.timeUntilAction.Before(p.timeUntilNavigation) {
		p.timeUntilAction = p.timeUntilNavigation
	}

	// Check if account does not exist/deleted
	if isPlayerNotExist(doc) {
		p.NotifyTelegram("player deleted or does not exist", false)
		return nil, errors.New("player deleted or does not exist")
	}

	// Check if banned
	if isBanned(doc) {
		p.NotifyTelegram("player banned", false)
		return nil, errors.New("player banned")
	}

	// Try again if clicked too fast!
	if isTooFast(doc) {
		r := getRandomInt(3123, 8765)
		log.Println("[" + p.nick + "]Clicked too fast! Sleeping for " + fmt.Sprintf("%.2f", float64(r)/1000) + "s and trying again...")
		time.Sleep(time.Duration(r) * time.Millisecond)
		return p.Navigate(path, action)
	}

	// Check if landed in anti-cheat check page
	if isAnticheatPage(doc) {
		res := p.solveAnticheat(doc)
		time.Sleep(MIN_WAIT_TIME - p.minRTT)
		if !res {
			log.Println("Anti cheat procedure failed...")
		}
		time.Sleep(MIN_WAIT_TIME - p.minRTT)
		return p.Navigate(path, action)
	}

	// Check if has new PMs
	if hasNewPM(doc) {
		m, err := p.getLastPM()
		if err != nil {
			panic(err)
		}

		// Ignore @sistema
		if m.from == "" {
			return p.Navigate(path, action)
		}

		// See telegram package - there is regex that MUST match below messages format in order to work
		// if m.moderator {
		// 	p.NotifyTelegram(fmt.Sprintf("Player '*%s' says: %s", m.from, m.text), false)
		// } else {
		// 	p.NotifyTelegram(fmt.Sprintf("Player '%s' says: %s", m.from, m.text), false)
		// }

		sender := m.from
		if m.moderator {
			sender = "*" + sender
		}

		log.Printf("Received PM from %s: %s", sender, m.text)
		telegram.SendMessage(fmt.Sprintf("Received PM from %s: %s", sender, m.text), false)

		generatedReply := getAIReply(m.text)
		if generatedReply == "" {
			generatedReply = "Nesupratau?"
		}

		time.Sleep(3 * time.Second)
		err = p.sendPM(m.from, generatedReply, doc)

		msg := fmt.Sprintf("Successfully sent PM reply to %s with generated text '%s'", sender, generatedReply)
		if err != nil {
			msg = fmt.Sprintf("Failed to send PM reply to %s with generated text '%s': %s", sender, generatedReply, err.Error())
		}

		log.Println(msg)
		telegram.SendMessage(msg, false)

		return p.Navigate(path, action)
	}

	return doc, nil
}

// Submit is used to submit forms in-game.
func (p *Player) Submit(path string, body io.Reader) (*goquery.Document, error) {
	timeNow := time.Now()

	p.manageBecomeOffline()

	// Wait between HTTP requests
	time.Sleep(p.timeUntilNavigation.Sub(timeNow))

	// Perform HTTP request and get response
	fullLink := p.fullLink(path)
	resp, err := p.httpRequest("POST", fullLink, body)
	if err != nil {
		log.Println("Failure occurred (#1): " + err.Error())
		log.Println("Sleeping for 5 seconds and trying again...")
		time.Sleep(5 * time.Second)
		return p.Submit(path, body)
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
		return p.Submit(path, body)
	}

	// Check if account does not exist/deleted
	if isPlayerNotExist(doc) {
		p.NotifyTelegram("player deleted or does not exist", false)
		return nil, errors.New("player deleted or does not exist")
	}

	// Check if banned
	if isBanned(doc) {
		p.NotifyTelegram("player banned", false)
		return nil, errors.New("player banned")
	}

	// Try again if clicked too fast!
	if isTooFast(doc) {
		r := getRandomInt(3123, 8765)
		log.Println("[" + p.nick + "]Clicked too fast! Sleeping for " + fmt.Sprintf("%.2f", float64(r)/1000) + "s and trying again...")
		time.Sleep(time.Duration(r) * time.Millisecond)
		return p.Submit(path, body)
	}

	// Mark wait time
	p.timeUntilNavigation = timeNow.Add(MIN_WAIT_TIME - p.minRTT)
	if p.timeUntilAction.Before(p.timeUntilNavigation) {
		p.timeUntilAction = p.timeUntilNavigation
	}

	// Check if landed in anti-cheat check page
	if isAnticheatPage(doc) {
		res := p.solveAnticheat(doc)
		time.Sleep(MIN_WAIT_TIME - p.minRTT)
		if !res {
			log.Println("Anti cheat procedure failed...")
		}
		return p.Submit(path, body)
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

func (p *Player) fullLink(path string) string {
	return p.rootAddress + strings.ReplaceAll(path, "{{ creds }}", "nick="+p.nick+"&pass="+p.pass)
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
