package main

/*

THIS PACKAGE IS ONLY USED TO GENERATE module/parduotuve ITEMS LIST.

1. Run below command while in "senasis amzius"
go run cmd/shop/main.go -nick aaaaa -pass aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa > shop.txt

2. Run the same command while in "naujasis amzius"
go run cmd/shop/main.go -nick aaaaa -pass aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa >> shop.txt

3. Sort & dedupe
sort -u shop.txt > shop2.txt

4. Copy contents from shop2.txt to module/parduotuve/parduotuve.go :)

*/

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
	"time"

	"tobot/config"
	"tobot/player"

	"github.com/PuerkitoBio/goquery"
)

var flagNick = flag.String("nick", "", "nick (taken from URL)")
var flagPass = flag.String("pass", "", "pass (taken from URL)")
var flagRoot = flag.String("root", "http://tob.lt", "Root URL of the website")
var flagUserAgent = flag.String("useragent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36", "User agent")
var flagMinRTT = flag.Duration("minrtt", 2*time.Millisecond, "Min RTT value")

func main() {
	flag.Parse()

	if *flagNick == "" || *flagPass == "" {
		panic("you need to specify nick and pass. See help with '-help'")
	}

	p := player.NewPlayer(&config.Player{
		Nick: *flagNick,
		Pass: *flagPass,
		Settings: config.Settings{
			RootAddress: flagRoot,
			UserAgent:   flagUserAgent,
			MinRTT:      flagMinRTT,
		},
	})

	// Open page containing list of categories
	doc, err := p.Navigate("/parda.php?{{ creds }}", false)
	if err != nil {
		panic(err)
	}

	// Iterate each category
	doc.Find("a[href*='nr='][href*='id=skyr']").Each(func(i int, s *goquery.Selection) {
		// Open category
		categoryLinkString, _ := s.Attr("href")
		categoryLinkURL, _ := url.Parse(categoryLinkString)
		categoryLinkURI := categoryLinkURL.RequestURI()
		categoryDoc, err := p.Navigate("/"+categoryLinkURI, false)
		if err != nil {
			panic(err)
		}

		// Iterate each subcategory in category
		categoryDoc.Find("a[href*='id=pard'][href*='page=']").Each(func(ii int, ss *goquery.Selection) {
			// Open subcategory
			subCategoryLinkString, _ := ss.Attr("href")
			subCategoryLinkURL, _ := url.Parse(subCategoryLinkString)
			subCategoryLinkURI := subCategoryLinkURL.RequestURI()
			subCategoryDoc, err := p.Navigate("/"+subCategoryLinkURI, false)
			if err != nil {
				panic(err)
			}

			// Iterate each item in subcategory
			subCategoryDoc.Find("a[href*='page='][href*='ka=']").Each(func(iii int, sss *goquery.Selection) {
				itemHref, _ := sss.Attr("href")
				itemHref = strings.SplitN(itemHref, "?", 2)[1]
				hrefPairs := strings.Split(itemHref, "&") // Yes, this is not &amp;
				var ka, page string
				for _, pair := range hrefPairs {
					pairPair := strings.Split(pair, "=")
					if pairPair[0] == "ka" {
						ka = pairPair[1]
					}
					if pairPair[0] == "page" {
						page = pairPair[1]
					}
				}
				fmt.Printf("	\"%s\":    \"%s\",\n", ka, page)
			})
		})

	})
}
