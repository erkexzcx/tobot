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
}

var rePM = regexp.MustCompile(`</a></b>:(.+)<br/>`)

func (p *Player) getLastPM() (*pm, error) {
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm", false)
	if err != nil {
		return p.getLastPM()
	}

	// Find element that contains first message
	s := doc.Find("div.got").First()

	// Return empty struct if last message received by "@SISTEMA"
	if s.Find("a:contains('[Atsakyti]')").Length() == 0 {
		return &pm{}, nil
	}

	// Find sender
	sender := s.Find("a[href*='zaisti.php'][href*='id=apie']").First().Text()
	if sender == "" {
		return nil, errors.New("unable to parse ")
	}
	sender = strings.TrimSpace(sender)

	// Check if moderator and if so - remove "*" from beginning
	moderator := false
	if strings.HasPrefix(sender, "*") {
		moderator = true
		sender = strings.TrimPrefix(sender, "*")
	}

	// Store RAW HTML for later use
	code, err := s.Html()
	if err != nil {
		return nil, err
	}

	match := rePM.FindStringSubmatch(code)
	if len(match) != 2 {
		log.Println("msg len:", len(match))
		return nil, errors.New("unable to parse message")
	}
	text := strings.TrimSpace(match[1])

	// Extract message
	return &pm{
		from:      sender,
		text:      text,
		moderator: moderator,
	}, nil
}

func (p *Player) sendPM(to, message string, doc *goquery.Document) error {
	path := "/meniu.php?{{ creds }}&id=siusti_pm&kam=" + to + "&ka="

	params := url.Values{}
	params.Add("zinute", message)
	params.Add("null", "Siųsti")
	body := strings.NewReader(params.Encode())

	// Need to wait in order to workaround "Palauk kelias sekundes ir bandykite vėl." error when sending
	sleepDuration := p.extractWaitTime(doc) - p.minRTT
	time.Sleep(sleepDuration)

	// Submit request
	_, err := p.Submit(path, body)
	return err
}

func (p *Player) handleScheduledReplies(doc *goquery.Document) {
	for {
		p.replyMux.Lock()
		isWaiting := p.waitingForReply
		p.replyMux.Unlock()

		if isWaiting {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		break
	}

	p.replyMux.Lock()
	defer p.replyMux.Unlock()
	for sendTo, message := range p.replyScheduled {
		err := p.sendPM(sendTo, message, doc)
		if err != nil {
			log.Fatalln(err)
			return
		}
		delete(p.replyScheduled, sendTo)
	}
}
