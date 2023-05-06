package player

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type pm struct {
	from      string
	text      string
	moderator bool
	system    bool
}

var (
	rePM             = regexp.MustCompile(`</a></b>:(.+)<br/>`)
	messageHTMLRegex = regexp.MustCompile(`(?i)<img[^>]+>`)
)

func (p *Player) getLastPM() (*pm, error) {
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm", false)
	if err != nil {
		return p.getLastPM()
	}

	// Find element that contains last received message
	s := doc.Find("div.got").First()

	// If sent by "@SISTEMA"
	if s.Find("a:contains('[Atsakyti]')").Length() == 0 {
		return &pm{system: true}, nil
	}

	// Find sender
	sender := s.Find("a[href*='zaisti.php'][href*='id=apie']").First().Text()
	if sender == "" {
		return nil, errors.New("unable to parse PM sender")
	}
	sender = strings.TrimSpace(sender)

	// Check if moderator and if so - remove "*" from beginning
	moderator := false
	if strings.HasPrefix(sender, "*") {
		moderator = true
		sender = strings.TrimPrefix(sender, "*")
	}

	// Extract message
	messageElementHTML, err := s.Html()
	if err != nil {
		return nil, err
	}
	match := rePM.FindStringSubmatch(messageElementHTML)
	if len(match) != 2 {
		log.Println("msg len:", len(match))
		return nil, errors.New("unable to parse PM message")
	}
	text := messageHTMLRegex.ReplaceAllString(strings.TrimSpace(match[1]), "")

	// Extract message
	return &pm{
		from:      sender,
		text:      text,
		moderator: moderator,
		system:    false,
	}, nil
}

func (p *Player) sendPM(to, message string, doc *goquery.Document) error {
	path := "/meniu.php?{{ creds }}&id=siusti_pm&kam=" + to + "&ka="

	params := url.Values{}
	params.Add("zinute", message)
	params.Add("null", "Siųsti")
	body := strings.NewReader(params.Encode())

	// Need to wait in order to workaround "Palauk kelias sekundes ir bandykite vėl." error when sending
	sleepDuration := p.extractWaitTime(doc) - *p.Config.Settings.MinRTT
	time.Sleep(sleepDuration)

	// Submit request
	_, err := p.Submit(path, body)
	return err
}

func (p *Player) dealWithPMs() error {
	// Get last PM
	lastPM, err := p.getLastPM()
	if err != nil {
		return err
	}

	// Open chat with sender
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm&ka="+lastPM.from, false)
	if err != nil {
		return err
	}

	// Get list of messages
	selection := doc.Find("div.got, div.sent")
}

/*
TODO - finish PM mechanism
TODO - review all those random functions
*/
