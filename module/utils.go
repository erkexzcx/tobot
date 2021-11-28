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

// Returns true if doc contains "Jūs pavargęs, bandykite vėl po keleto sekundžių.."
func IsActionTooFast(doc *goquery.Document) bool {
	return doc.Find("div:contains('Jūs pavargęs, bandykite vėl po keleto sekundžių..')").Length() > 0
}
