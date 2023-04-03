package player

import (
	"crypto/md5"
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

func (p *Player) solveAnticheat(doc *goquery.Document) bool {
	// Get map of color -> linkToClick
	colorToLinkMap, ok := getColorToLinkMap(p, doc)
	if !ok {
		log.Println("Failed anticheat main #1")
		return false
	}

	// Find which color we should click, according to each color's MD5 checksum.\
	//
	// Fun fact: re-downloading captcha image gives the same color text, but formatted differently,
	// therefore keep refreshing until you successfully read it.
	for i := 0; i < getRandomInt(20, 30); i++ {

		// Find color name
		color := getColorToClickName(p, doc)
		if color == "" {
			continue
		}
		if color == "nieko" {
			log.Println("Anti-cheat ran out of time and failed")
			return false
		}

		// Click the color
		resp, err := p.httpRequest("GET", colorToLinkMap[color], nil)
		if err != nil {
			log.Println("Failed anticheat main #2:", err)
			return false
		}
		defer resp.Body.Close()
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println("Failed anticheat main #3:", err)
			return false
		}

		// Return status
		return doc.Find("div:contains('Galite žaisti toliau.')").Length() > 0
	}
	log.Println("Anti-cheat ran out of time and failed")
	return false
}

func getColorToClickName(p *Player, doc *goquery.Document) string {
	// Extract link of captcha
	src, found := doc.Find("img[src*='spalva.php']").Attr("src")
	if !found {
		log.Println("Failed anticheat img #1: image attribute 'src' not found")
		return ""
	}
	parsedSrc, err := url.Parse(src)
	if err != nil {
		log.Println("Failed anticheat img #2:", err)
		return ""
	}
	imageLink := p.rootAddress + "/" + parsedSrc.RequestURI()

	// Download captcha image
	resp, err := p.httpRequest("GET", imageLink, nil)
	if err != nil {
		log.Println("Failed anticheat img #3:", err)
		return ""
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed anticheat img #4:", err)
		return ""
	}

	// Read text from image
	tessClient.SetImageFromBytes(content)
	text, err := tessClient.Text()
	if err != nil {
		log.Fatalln("Failed anticheat img #5:", err)
	}
	colorText := strings.ToLower(reColor.ReplaceAllString(text, "")) // Already trimmed by regex

	// If it's too late
	if colorText == "nieko" {
		return "nieko"
	}

	// Find color text
	for _, v := range MD5SumToColor {
		if colorText == v {
			return v
		}
	}
	return ""
}

func getColorToLinkMap(p *Player, doc *goquery.Document) (map[string]string, bool) {
	failed := false

	colorToLinkMap := make(map[string]string)

	// Download each image and save to file + generate & print MD5
	doc.Find("img[src*='antibotimg.php'][src*='nr=']").Each(func(i int, s *goquery.Selection) {
		if failed {
			return
		}

		// Get image URL
		src, found := s.Attr("src")
		if !found {
			log.Println("Failed anticheat #1: image attribute 'src' not found")
			failed = true
			return
		}
		parsedSrc, err := url.Parse(src)
		if err != nil {
			log.Println("Failed anticheat #2:", err)
			failed = true
			return
		}
		imageLink := p.rootAddress + "/" + parsedSrc.RequestURI()

		// Download image
		resp, err := p.httpRequest("GET", imageLink, nil)
		if err != nil {
			log.Println("Failed anticheat #3:", err)
			failed = true
			return
		}
		defer resp.Body.Close()
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed anticheat #4:", err)
			failed = true
			return
		}

		// Get MD5 checksum of image bytes
		md5sum := fmt.Sprintf("%x", md5.Sum(content))

		// Find such color
		color, ok := MD5SumToColor[md5sum]
		if !ok {
			log.Println("Failed anticheat #5: color not found (" + md5sum + ")")
			failed = true
			return
		}

		// Find click link
		href, found := s.Parent().Attr("href")
		if !found {
			log.Println("Failed anticheat #6: a attribute 'href' not found")
			failed = true
			return
		}
		parsedHref, err := url.Parse(href)
		if err != nil {
			log.Println("Failed anticheat #7:", err)
			failed = true
			return
		}
		aLink := p.rootAddress + "/" + parsedHref.RequestURI()

		// Add to map
		colorToLinkMap[color] = aLink
	})

	return colorToLinkMap, !failed
}

func init() {
	// Init tesseract OCR
	tessClient = gosseract.NewClient()
	//defer tessClient.Close()

	tessClient.SetLanguage("lit")
}
