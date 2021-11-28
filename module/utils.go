package module

import (
	"log"

	"github.com/PuerkitoBio/goquery"
)

func DumpHTML(doc *goquery.Document) {
	html, err := doc.Html()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(html)
}

// Returns true if doc contains "Jūs pavargęs, bandykite vėl po keleto sekundžių"-alike message
func IsActionTooFast(doc *goquery.Document) bool {
	if doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length() > 0 {
		return true
	}
	if doc.Find("div:contains('Bandykite po kelių sekundžių, pavargote.')").Length() > 0 {
		return true
	}
	return false
}
