package main

/*

THIS PACKAGE IS USED TO MASS SEND PMs to given list of players (from the file)

1. create file players.txt like this:
player1
player2
dariusltu
omexgaa
anotherplayer

2. Run the command
go run cmd/pm/main.go -nick aaaaa -pass aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa -accounts players.txt -message 'hello world!'

*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"tobot/player"
)

var flagNick = flag.String("nick", "", "nick (taken from URL)")
var flagPass = flag.String("pass", "", "pass (taken from URL)")
var flagRoot = flag.String("root", "http://tob.lt", "root URL (don't change unless you know what you are doing)")
var flagAccounts = flag.String("accounts", "players.txt", "file containing list of nicknames (newline separated)")
var flagMessage = flag.String("message", "", "Message to send to the users")

func main() {
	flag.Parse()

	if *flagNick == "" || *flagPass == "" {
		panic("you need to specify nick and pass. See help with '-help'")
	}

	p := player.NewPlayer(
		*flagNick,
		*flagPass,
		*flagRoot,
		"tob.lt",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36",
		0,
		nil,
		0,
		0,
		0,
		0,
		0,
		0,
	)

	accountsString, err := ioutil.ReadFile(*flagAccounts)
	if err != nil {
		panic(err)
	}
	accounts := strings.Split(string(accountsString), "\n")
	for _, account := range accounts {
		account = strings.TrimSpace(account)
		if account == "" {
			continue
		}

		path := "/meniu.php?{{ creds }}&id=siusti_pm&kam=" + account + "&ka="
		params := url.Values{}
		params.Add("zinute", *flagMessage)
		params.Add("null", "Siųsti")
		body := strings.NewReader(params.Encode())
		doc, err := p.Submit(path, body)
		if err != nil {
			panic(err)
		}

		if doc.Find("div:contains('Išsiųsta')").Length() == 0 {
			html, _ := doc.Html()
			fmt.Println(html)
			log.Fatalln("no messages been sent to", account)
		}

		fmt.Println("message sent to:", account)
		time.Sleep(5 * time.Second)
	}
}
