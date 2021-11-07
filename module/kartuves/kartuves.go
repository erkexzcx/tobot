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
			newPattern := strings.ReplaceAll(pattern, "_", letter)
			db.Exec("UPDATE data SET pattern=?, remaining=? WHERE ? LIKE pattern", newPattern, "", newPattern)
			return &module.Result{CanRepeat: false, Error: nil}
		}
		if tmpDoc.Find("div:contains('Jus pakartas')").Length() > 0 {
			log.Println("Jus pakartas!")

			remainingLettersSlice := make([]string, 0, len(remainingLetters))
			for l := range remainingLetters {
				remainingLettersSlice = append(remainingLettersSlice, l)
			}
			remainingLettersString := strings.Join(remainingLettersSlice, "")
			remainingLettersString = strings.ReplaceAll(remainingLettersString, letter, "")
			db.Exec("UPDATE data SET pattern=?, remaining=? WHERE ? LIKE pattern", pattern, remainingLettersString, pattern)

			return &module.Result{CanRepeat: false, Error: nil}
		}

		module.DumpHTML(tmpDoc)
		return &module.Result{CanRepeat: false, Error: nil}
	}

	// I, A and S letters are the most popular ones, so if it exists - must press it
	if _, found := remainingLetters["I"]; found {
		return clickLetter("I")
	}
	if _, found := remainingLetters["A"]; found {
		return clickLetter("A")
	}
	if _, found := remainingLetters["S"]; found {
		return clickLetter("S")
	}

	// Find pattern
	var selectedPattern, selectedRemaining string
	err = db.QueryRow("SELECT pattern, remaining FROM data WHERE ? LIKE pattern LIMIT 1", pattern).Scan(&selectedPattern, &selectedRemaining)

	// If no pattern found
	if errors.Is(err, sql.ErrNoRows) {
		remainingLettersSlice := make([]string, 0, len(remainingLetters))
		for l := range remainingLetters {
			remainingLettersSlice = append(remainingLettersSlice, l)
		}
		remainingLettersString := strings.Join(remainingLettersSlice, "")
		db.Exec("INSERT INTO data(pattern, remaining) values(?, ?)", pattern, remainingLettersString)
		for k := range remainingLetters {
			return clickLetter(k)
		}
		panic("This should not happen")
	}

	// If word is fully known (pattern, but without '_' symbols)
	if selectedRemaining == "" {
		selectedPatternSlice := strings.Split(selectedPattern, "")
		for _, l := range selectedPatternSlice {
			if _, ok := remainingLetters[l]; ok {
				return clickLetter(l)
			}
		}
	}

	// Default - pattern is found
	selectedPatternSlice := strings.Split(selectedPattern, "")
	for _, l := range selectedPatternSlice {
		if l == "_" {
			continue
		}
		if _, ok := remainingLetters[l]; ok {
			return clickLetter(l)
		}
	}
	selectedRemainingSlice := strings.Split(selectedRemaining, "")
	for _, srLetter := range selectedRemainingSlice {
		if _, ok := remainingLetters[srLetter]; ok {
			return clickLetter(srLetter)
		}
	}

	// Should not reach this, but click on any letter anyway...
	for k := range remainingLetters {
		return clickLetter(k)
	}
	return nil
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
