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
