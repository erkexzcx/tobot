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
			db.Exec("INSERT OR IGNORE INTO tried(word, ok) values(?, 1)", strings.ReplaceAll(pattern, "_", letter))
			return &module.Result{CanRepeat: false, Error: nil}
		}
		if tmpDoc.Find("div:contains('Jus pakartas')").Length() > 0 {
			log.Println("Jus pakartas!")
			db.Exec("INSERT OR IGNORE INTO tried(word, ok) values(?, 0)", pattern)
			return &module.Result{CanRepeat: false, Error: nil}
		}

		module.DumpHTML(tmpDoc)
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// Find all results matching required pattern and find the most popular letter in them
	letters := make(map[string]int, 0)
	rows, err := db.Query("SELECT word FROM words WHERE word LIKE ?", pattern)
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}
	defer rows.Close()
	for rows.Next() {
		var word string
		err = rows.Scan(&word)
		if err != nil {
			return &module.Result{CanRepeat: false, Error: err}
		}
		for _, letter := range strings.Split(word, "") {
			if _, remainingLetterFound := remainingLetters[letter]; remainingLetterFound {
				if _, found := letters[letter]; found {
					letters[letter]++
				} else {
					letters[letter] = 1
				}
			}
		}
	}
	log.Println(letters)
	err = rows.Err()
	if err != nil {
		return &module.Result{CanRepeat: false, Error: err}
	}

	// If no letters were found (e.g. no results from the query), just click on the first one...
	if len(letters) == 0 {
		log.Printf("No query results found for pattern '%s'...\n", pattern)
		for k := range remainingLetters {
			return clickLetter(k)
		}
		panic("This should not happen")
	}

	// Find the most popular letter from the map and click on it
	var mostPopular string
	var mostPopularCount int
	for k, v := range letters {
		if v > mostPopularCount {
			mostPopularCount = v
			mostPopular = k
		}
	}
	log.Printf("clicking on most popular letter '%s' (occurances=%d) for pattern '%s'...\n", mostPopular, mostPopularCount, pattern)
	return clickLetter(mostPopular)
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
	db, err = sql.Open("sqlite3", "file:./kartuves.db?_mutex=full")
	if err != nil {
		panic(err)
	}
	// db.Close()

	module.Add("kartuves", &Kartuves{})
}
