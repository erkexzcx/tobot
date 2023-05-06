package player

import (
	"log"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"
	"tobot/comms"

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
	rePmSent         = regexp.MustCompile(`</a></b> - (.+)<br/>`)   // Class "send"
	rePmGot          = regexp.MustCompile(`</a></b>: (.+)<br/><i>`) // Class "got"
	messageHTMLRegex = regexp.MustCompile(`(?i)<img[^>]+>`)
)

func parsePmHtml(s *goquery.Selection) *pm {
	// If sent by "@SISTEMA"
	if s.Find("b:contains('» @SISTEMA')").Length() > 0 {
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
	var match []string
	if pm.received {
		match = rePmGot.FindStringSubmatch(messageElementHTML)
	} else {
		match = rePmSent.FindStringSubmatch(messageElementHTML)
	}
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

func (p *Player) sendPM(to, message string, doc *goquery.Document) error {
	path := "/meniu.php?{{ creds }}&id=siusti_pm&kam=" + to + "&ka="

	params := url.Values{}
	params.Add("zinute", message)
	params.Add("null", "Siųsti")
	body := strings.NewReader(params.Encode())

	// Submit request
	_, err := p.Submit(path, body)
	return err
}

func (p *Player) dealWithPMs() error {
	// Get last PM
	lastPM := p.getLastReceivedPM()

	// Ignore if system message
	if lastPM.system {
		return nil
	}

	// Format for logs
	modifiedNick := lastPM.nick
	if lastPM.moderator {
		modifiedNick = "*" + modifiedNick
	}

	log.Printf("Received PM from %s: %s\n", modifiedNick, lastPM.text)
	comms.ForwardMessageToTelegram(lastPM.text, modifiedNick, true)

	// Open chat only with sender
	doc, err := p.Navigate("/meniu.php?{{ creds }}&id=pm&ka="+lastPM.nick, false)
	if err != nil {
		return err
	}

	// Get slice of messages. This contains messages from the latest to the oldest one.
	allSendersPMs := []*pm{}
	doc.Find("div.got, div.send").Each(func(i int, s *goquery.Selection) {
		allSendersPMs = append(allSendersPMs, parsePmHtml(s))
	})

	// Reverse the order of the messages slice, so it would be from the oldest to the latest one.
	for i, j := 0, len(allSendersPMs)-1; i < j; i, j = i+1, j-1 {
		allSendersPMs[i], allSendersPMs[j] = allSendersPMs[j], allSendersPMs[i]
	}

	// Print messages for debug
	openaiMsgs := []*comms.OpenaiMessage{}
	for _, p := range allSendersPMs {
		openaiMsg := &comms.OpenaiMessage{
			Received: p.received,
			Message:  p.text,
		}
		openaiMsgs = append(openaiMsgs, openaiMsg)
	}

	// Get reply from openai api
	openaiReply := comms.GetOpenAIReply(openaiMsgs...)

	// Sleep according to amount of symbols within the reply (to simulate user writing)
	sleepDuration := CalculateSleepTime(openaiReply, 30)
	time.Sleep(sleepDuration)

	// Send message back to user
	for {
		err = p.sendPM(lastPM.nick, openaiReply, doc)
		if err == nil {
			break
		}
		log.Println("Failed to send PM, retrying...")
		comms.SendMessageToTelegram("Failed to send PM (" + err.Error() + "), retrying...")
	}

	log.Printf("AI Replied to %s: %s\n", modifiedNick, openaiReply)
	comms.ForwardMessageToTelegram(openaiReply, modifiedNick, false)

	return nil
}

// CalculateSleepTime calculates the time to sleep based on the input text and average typing speed
func CalculateSleepTime(text string, wpm float64) time.Duration {
	avgWordLength := 4.7
	cpm := wpm * avgWordLength
	chars := len(strings.TrimSpace(text))
	secondsPerChar := 60.0 / cpm
	sleepSeconds := float64(chars) * secondsPerChar
	return time.Duration(math.Round(sleepSeconds * 1e9))
}
