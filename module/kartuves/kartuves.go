package kartuves

import (
	"database/sql"
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
	"tobot/module"
	"tobot/player"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

var (
	reExtractWait    = regexp.MustCompile(`Zaisti galesite po (\d+\w)`)
	reExtractPattern = regexp.MustCompile(`(( _ |[A-Z]){5,})<br\/>`)
)

var db *sql.DB

const dbQueryFirstMatch = `SELECT word FROM words WHERE word LIKE ? LIMIT 1`
const dbQueryCount = `SELECT COUNT(word) FROM words WHERE word LIKE ?`

type Kartuves struct{}

func (obj *Kartuves) Validate(settings map[string]string) error {
	for k := range settings {
		if strings.HasPrefix(k, "_") {
			continue
		}
		return errors.New("unrecognized key '" + k + "'")
	}
	return nil
}

func (obj *Kartuves) Perform(p *player.Player, settings map[string]string) *module.Result {
	path := "/kartuves.php?{{ creds }}"

	// Download page
	doc, err := p.Navigate(path, false)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// Check if we can play
	if doc.Find("div:contains('Zaisti kartuves')").Length() > 0 {
		return obj.Perform(p, settings)
	}

	// Check if we have to wait
	if doc.Find("div:contains('Zaisti galesite po')").Length() > 0 {
		log.Println("waiting for next game...")
		waitUntilGame(doc)
		return obj.Perform(p, settings)
	}

	// Check if we are in the right page
	if doc.Find("div:contains('Spekite raide:')").Length() == 0 {
		module.DumpHTML(doc)
		return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred - failed to understand where we are")}
	}

	// Get RAW HTML
	html, err := doc.Html()
	if err != nil {
		return &module.Result{CanRepeat: false, Error: errors.New("failed to retrieve HTML")}
	}

	// Extract pattern
	matches := reExtractPattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		return &module.Result{CanRepeat: false, Error: errors.New("failed to extract pattern")}
	}
	pattern := strings.ReplaceAll(matches[1], " ", "")

	// Extract remaining letters
	remainingLetters := map[string]string{} // map[letter]link
	doc.Find("a[href*='&id=speti&ka=']").Each(func(i int, s *goquery.Selection) {
		letter := strings.TrimSpace(s.Text())

		href, _ := s.Attr("href")
		hrefURL, err := url.Parse(href)
		if err != nil {
			log.Println(err)
			return
		}
		letterPath := hrefURL.RequestURI()

		remainingLetters[letter] = letterPath
	})
	if len(remainingLetters) == 0 {
		return &module.Result{CanRepeat: false, Error: errors.New("no remaining letters found")}
	}

	clickLetter := func(letter string) *module.Result {
		tmpDoc, err := p.Navigate(remainingLetters[letter], false)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		if tmpDoc.Find("div:contains('Tokios raides zodyje nera.')").Length() > 0 {
			return &module.Result{CanRepeat: true, Error: nil}
		}
		if tmpDoc.Find("div:contains('Atspejote raide')").Length() > 0 {
			return &module.Result{CanRepeat: true, Error: nil}
		}
		if tmpDoc.Find("div:contains('Atspejote visa zodi!')").Length() > 0 {
			log.Println("Zodis atspetas!")
			return &module.Result{CanRepeat: true, Error: nil}
		}
		if tmpDoc.Find("div:contains('Jus pakartas')").Length() > 0 {
			log.Println("Jus pakartas!")
			return &module.Result{CanRepeat: true, Error: nil}
		}

		module.DumpHTML(tmpDoc)
		return &module.Result{CanRepeat: false, Error: nil}
	}

	matchesInDB := matchesInDatabase(pattern)
	if matchesInDB == 0 {
		// Click first letter in remaining letters list
		for k := range remainingLetters {
			return clickLetter(k)
		}
	}

	if _, f := remainingLetters["I"]; f {
		return clickLetter("I")
	}
	if _, f := remainingLetters["A"]; f {
		return clickLetter("A")
	}
	if _, f := remainingLetters["S"]; f {
		return clickLetter("S")
	}
	if _, f := remainingLetters["E"]; f {
		return clickLetter("E")
	}
	if _, f := remainingLetters["T"]; f {
		return clickLetter("T")
	}
	if _, f := remainingLetters["N"]; f {
		return clickLetter("N")
	}

	firstRes := findFirstMatchInDatabase(pattern)
	for _, letter := range strings.Split(firstRes, "") {
		if _, f := remainingLetters[letter]; f {
			return clickLetter(letter)
		}
	}
	// If no letters left in matches
	for letter := range remainingLetters {
		return clickLetter(letter)
	}

	module.DumpHTML(doc)
	return &module.Result{CanRepeat: false, Error: errors.New("unknown error occurred")}
}

func findFirstMatchInDatabase(s string) string {
	var res string
	err := db.QueryRow(dbQueryFirstMatch, s).Scan(&res)
	if err != nil {
		panic(err)
	}
	return res
}

func matchesInDatabase(s string) int {
	var count int
	err := db.QueryRow(dbQueryCount, s).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count
}

func waitUntilGame(doc *goquery.Document) {
	html, err := doc.Html()
	if err != nil {
		log.Println(err)
		return
	}

	matches := reExtractWait.FindStringSubmatch(html)
	if len(matches) != 2 {
		log.Println("failed to extract wait time")
		return
	}

	d, err := time.ParseDuration(matches[1])
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(d + (time.Second / 2))
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "file:./kartuves.db")
	if err != nil {
		panic(err)
	}
	// db.Close()

	module.Add("kartuves", &Kartuves{})
}
