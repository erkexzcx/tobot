package player

import (
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type pm struct {
	received  bool
	nick      string
	text      string
	moderator bool
	system    bool
}

var (
	rePM             = regexp.MustCompile(`</a></b>:(.+)<br/>`)
	messageHTMLRegex = regexp.MustCompile(`(?i)<img[^>]+>`)
)

func parsePmHtml(s *goquery.Selection) *pm {
	// If sent by "@SISTEMA"
	if s.Find("a:contains('[Atsakyti]')").Length() == 0 {
		return &pm{system: true}
	}

	// Create new PM object
	pm := &pm{}

	// Check if message was received or sent
	pm.received = s.HasClass("got")

	// Parse message player nick who sent or who received message from us
	pm.nick = s.Find("a[href*='zaisti.php'][href*='id=apie']").First().Text()
	pm.nick = strings.TrimSpace(pm.nick)
	if pm.nick == "" {
		log.Fatalln("Unable to parse PM pm.nick")
	}

	// If moderator - remove * char from their nick
	if strings.HasPrefix(pm.nick, "*") {
		pm.nick = strings.TrimPrefix(pm.nick, "*")
		pm.moderator = true
	}

	// Parse message text
	messageElementHTML, err := s.Html()
	if err != nil {
		return nil
	}
	match := rePM.FindStringSubmatch(messageElementHTML)
	if len(match) != 2 {
		log.Fatalln("Unable to parse PM message text")
	}
	pm.text = match[1]
	pm.text = strings.TrimSpace(pm.text)
	pm.text = messageHTMLRegex.ReplaceAllString(pm.text, "")

	return pm
}

func (p *Player) getLastReceivedPM() *pm {
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm", false)
	if err != nil {
		return p.getLastReceivedPM()
	}

	s := doc.Find("div.got").First()

	return parsePmHtml(s)
}

// func (p *Player) sendPM(to, message string, doc *goquery.Document) error {
// 	path := "/meniu.php?{{ creds }}&id=siusti_pm&kam=" + to + "&ka="

// 	params := url.Values{}
// 	params.Add("zinute", message)
// 	params.Add("null", "Siųsti")
// 	body := strings.NewReader(params.Encode())

// 	// Need to wait in order to workaround "Palauk kelias sekundes ir bandykite vėl." error when sending
// 	sleepDuration := p.extractWaitTime(doc) - *p.Config.Settings.MinRTT
// 	time.Sleep(sleepDuration)

// 	// Submit request
// 	_, err := p.Submit(path, body)
// 	return err
// }

func (p *Player) dealWithPMs() error {
	// Get last PM
	lastPM := p.getLastReceivedPM()

	// Ignore if system message
	if lastPM.system {
		return nil
	}

	// Open chat only with sender
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm&ka="+lastPM.nick, false)
	if err != nil {
		return err
	}

	// Get slice of messages. This contains messages from the latest to the oldest one.
	allSendersPMs := []*pm{}
	doc.Find("div.got, div.sent").Each(func(i int, s *goquery.Selection) {
		allSendersPMs = append(allSendersPMs, parsePmHtml(s))
	})

	// Reverse the order of the messages slice, so it would be from the oldest to the latest one.
	for i, j := 0, len(allSendersPMs)-1; i < j; i, j = i+1, j-1 {
		allSendersPMs[i], allSendersPMs[j] = allSendersPMs[j], allSendersPMs[i]
	}

	// Print messages for debug
	for _, pm := range allSendersPMs {
		log.Println(pm)
	}

	return nil
}
