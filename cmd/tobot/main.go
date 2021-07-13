package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
	"tobot"

	"tobot/config"
	"tobot/player"

	_ "tobot/module/all"

	"github.com/PuerkitoBio/goquery"
	tb "gopkg.in/tucnak/telebot.v2"
	"gopkg.in/yaml.v2"
)

var configPath = flag.String("config", "config.yml", "path to config file")
var activitiesDir = flag.String("activities", "activities", "path to activities directory")
var shopFlag = flag.Bool("shop", false, "extracts all item codes and their shop pages from the main in-game shop")

func main() {
	flag.Parse()

	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	a := make([]*tobot.Activity, 0)
	files, err := filepath.Glob(*activitiesDir + "/*.yml")
	if err != nil {
		panic("Failed to read activities .yml files: " + err.Error())
	}
	for _, f := range files {
		if strings.HasPrefix(path.Base(f), "_") {
			continue // Skip '_*.yml' files
		}
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalln(err)
		}

		var aa *tobot.Activity
		if err := yaml.Unmarshal(contents, &aa); err != nil {
			log.Fatalln(err)
		}
		a = append(a, aa)
	}

	telegramBot, err := tb.NewBot(tb.Settings{
		Token:  c.TelegramApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
		return
	}

	parsedLink, err := url.Parse(c.RootAddress)
	if err != nil {
		log.Fatalln(err)
		return
	}

	becomeOfflineEvery := strings.Split(c.BecomeOfflineEvery, ",")
	becomeOfflineFor := strings.Split(c.BecomeOfflineFor, ",")

	becomeOfflineEveryFrom, _ := time.ParseDuration(becomeOfflineEvery[0])
	becomeOfflineEveryTo, _ := time.ParseDuration(becomeOfflineEvery[1])
	becomeOfflineForFrom, _ := time.ParseDuration(becomeOfflineFor[0])
	becomeOfflineForTo, _ := time.ParseDuration(becomeOfflineFor[1])

	ps := &player.PlayerSettings{
		Nick: c.Nick,
		Pass: c.Pass,

		MinRTTTime: c.MinRTTTime - (1 * time.Millisecond),

		TelegramBot:  telegramBot,
		TelegramChat: &tb.Chat{ID: c.TelegramChatId},

		RootLink: c.RootAddress,

		HeaderUserAgent: c.UserAgent,
		HeaderHost:      parsedLink.Host,

		BecomeOffline: c.BecomeOffline,

		BecomeOfflineEveryFrom: becomeOfflineEveryFrom,
		BecomeOfflineEveryTo:   becomeOfflineEveryTo,

		BecomeOfflineForFrom: becomeOfflineForFrom,
		BecomeOfflineForTo:   becomeOfflineForTo,
	}
	p := player.NewPlayer(ps)

	if *shopFlag {
		parseShopItems(p)
		return
	}

	tobot.Start(p, c, a)
}

func parseShopItems(p *player.Player) {
	fmt.Println("var itemsPagesMap = map[string]string{")

	doc, err := p.Navigate("/parda.php?{{ creds }}", false)
	if err != nil {
		panic(err)
	}

	doc.Find("a[href*='parda.php?'][href*='id=skyr']").Each(func(i int, s *goquery.Selection) {
		href, found := s.Attr("href")
		if !found {
			panic("'href' attr not found")
		}
		link, err := url.Parse(href)
		if err != nil {
			panic(err)
		}

		doc2, err := p.Navigate("/"+link.RequestURI(), false)
		if err != nil {
			panic(err)
		}

		doc2.Find("a[href*='id=pard'][href*='page=']").Each(func(i int, s *goquery.Selection) {
			href2, found2 := s.Attr("href")
			if !found2 {
				panic("'href' attr not found")
			}
			link2, err := url.Parse(href2)
			if err != nil {
				panic(err)
			}

			doc3, err := p.Navigate("/"+link2.RequestURI(), false)
			if err != nil {
				panic(err)
			}

			doc3.Find("a[href*='id=parduot'][href*='ka=']").Each(func(i int, s *goquery.Selection) {
				href3, found3 := s.Attr("href")
				if !found3 {
					panic("'href' attr not found")
				}
				link3, err := url.Parse(href3)
				if err != nil {
					panic(err)
				}

				query, err := url.ParseQuery(link3.RawQuery)
				if err != nil {
					panic(err)
				}

				ka, f1 := query["ka"]
				page, f2 := query["page"]
				if !f1 || !f2 {
					panic("Either 'ka' or 'page' keys not found in URL...")
				}

				fmt.Printf("\t\"%s\": \"%s\",\n", ka[0], page[0])
			})
		})
	})

	fmt.Println("}")
}
