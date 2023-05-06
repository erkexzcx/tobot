package player

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/otiai10/gosseract"
)

var tessClient *gosseract.Client

var MD5SumToColor = map[string]string{
	"00b2bd826148fac618dd782570dde345": "raudona",
	"06a957d0d6eddd395a2bff5210294950": "melyna",
	"422bd9e530718e8c2f7ecfab43762f28": "oranzine",
	"f03b55e7ea057ddd1f6c6ab1335f7183": "rozine",
	"b51e5d380e02ecb9922ec8d5494be7a9": "zalia",
	"61d8c65d185686d3af8419292d37667b": "juoda",
	"ce5118cb6325b96831eb91714815b8ef": "geltona",
	"d0da1f6f707eeda3641430a6e94ee91c": "violetine",
	"32f793aa2620d2be93d2573ccff75b75": "ruda",
}

var reColor = regexp.MustCompile(`[^a-zA-Z]`)

func (p *Player) solveAnticheat(doc *goquery.Document) error {
	// From the goquery Document, get matrix of clickable colours in a
	// form of map[colorName]linkToClickForThatColor...
	colorClickableMatrix, err := getColorClickableMatrix(p, doc)
	if err != nil {
		return err
	}

	// Read captcha image to understand which color we have to click
	color, err := getColorToClickName(p, doc)
	if err != nil {
		return err
	}

	// Click on the found colour
	doc, err = p.Navigate(colorClickableMatrix[color], false)
	if err != nil {
		return errors.New("Failed to click color " + color + ": " + err.Error())
	}

	// Check if we passed the anti-cheat
	success := doc.Find("div:contains('Galite žaisti toliau.')").Length() > 0
	if success {
		return nil
	}

	// Check if something went wrong
	contents, _ := doc.Html()
	log.Println(contents)
	return errors.New("Failed to pass anti-cheat due unknown reason (color: " + color + ")")
}

// Fun fact: re-downloading captcha image gives the same color text, but formatted differently,
// therefore keep refreshing until you successfully read it.
func getColorToClickName(p *Player, doc *goquery.Document) (string, error) {
	// Extract link of captcha
	src, found := doc.Find("img[src*='spalva.php']").Attr("src")
	if !found {
		return "", errors.New("failed to find captcha image (src attribute not found)")
	}

	// Get image link
	parsedSrc, err := url.Parse(src)
	if err != nil {
		return "", errors.New("Failed to parse captcha image link: " + err.Error())
	}
	imageLink := *p.Config.Settings.RootAddress + "/" + parsedSrc.RequestURI()

	// At max 100 image captcha refreshes...
	for i := 0; i < 100; i++ {
		// Get captcha image response
		resp, err := p.httpRequest("GET", imageLink, nil)
		if err != nil {
			return "", errors.New("Failed to get captcha image response: " + err.Error())
		}
		defer resp.Body.Close()

		// Download image body
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.New("Failed to read captcha image body: " + err.Error())
		}

		// Read text from image
		tessClient.SetImageFromBytes(content)
		text, err := tessClient.Text()
		if err != nil {
			return "", errors.New("Failed to read text from captcha image: " + err.Error())
		}
		colorText := strings.ToLower(reColor.ReplaceAllString(text, "")) // Already trimmed by regex

		// Check if we ran out of time
		if colorText == "nieko" {
			return "", errors.New("anti-cheat ran out of time and failed")
		}

		// Check if returned color text is recognized
		for _, v := range MD5SumToColor {
			if colorText == v {
				return v, nil
			}
		}

		// If not returned - try the same again in the next loop...
	}

	return "", errors.New("reached too many tries to read color from the captcha image")
}

// map[colorName]linkToClick
func getColorClickableMatrix(p *Player, doc *goquery.Document) (map[string]string, error) {
	colorNameToLink := make(map[string]string)
	var returnableError error = nil
	doc.Find("img[src*='antibotimg.php'][src*='nr=']").Each(func(i int, s *goquery.Selection) {
		// Extract image URL
		src, found := s.Attr("src")
		if !found {
			returnableError = errors.New("failed to find clickable captcha image link (src attribute not found)")
			return
		}

		// Get image link
		parsedSrc, err := url.Parse(src)
		if err != nil {
			returnableError = errors.New("Failed to parse clickable captcha image link URL: " + err.Error())
			return
		}
		imageLink := *p.Config.Settings.RootAddress + "/" + parsedSrc.RequestURI()

		// Download image response
		resp, err := p.httpRequest("GET", imageLink, nil)
		if err != nil {
			returnableError = errors.New("Failed to get clickable captcha image response: " + err.Error())
			return
		}
		defer resp.Body.Close()

		// Download image body
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			returnableError = errors.New("Failed to read clickable captcha image body: " + err.Error())
			return
		}

		// Get MD5 checksum of image bytes
		md5sum := fmt.Sprintf("%x", md5.Sum(content))

		// Find such color
		color, ok := MD5SumToColor[md5sum]
		if !ok {
			returnableError = errors.New("Failed to find color by MD5 checksum: " + md5sum)
			return
		}

		// Find click link
		href, found := s.Parent().Attr("href")
		if !found {
			returnableError = errors.New("failed to find clickable captcha image link (a attribute 'href' not found)")
			return
		}

		// Create image click link and add to the map
		parsedHref, err := url.Parse(href)
		if err != nil {
			returnableError = errors.New("Failed to parse clickable captcha image link URL: " + err.Error())
			return
		}
		colorNameToLink[color] = "/" + parsedHref.RequestURI() // For some reasons it misses "/" at the beginning of the link
	})

	return colorNameToLink, returnableError
}

func init() {
	// Init tesseract OCR
	tessClient = gosseract.NewClient()
	//defer tessClient.Close()

	tessClient.SetLanguage("lit")
}
