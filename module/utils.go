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
	values := []string{
		"Jūs pavargęs, bandykite vėl po keleto sekundžių..",
		"Bandykite po kelių sekundžių, pavargote.",
	}
	for _, s := range values {
		if doc.Find("div:contains('"+s+"')").Length() > 0 {
			return true
		}
	}
	return false
}

// Returns true if doc contains "Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!"-alike message
func IsInvalidClick(doc *goquery.Document) bool {
	values := []string{
		"Taip negalima! turite eiti atgal ir vėl bandyti atlikti veiksmą!",
		"Taip negalima! turite eiti atgal ir vėl pulti!",
	}
	for _, s := range values {
		if doc.Find("div:contains('"+s+"')").Length() > 0 {
			return true
		}
	}
	return false
}
