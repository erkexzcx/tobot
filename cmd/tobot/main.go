package main

import (
	"flag"
	"log"
	"tobot"

	"tobot/comms"
	"tobot/config"
	"tobot/player"

	_ "tobot/module/all"
)

var configPath = flag.String("config", "config.yml", "path to config file")

func main() {
	flag.Parse()

	// Parse configuration file
	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Start comms package
	comms.InitComms(c)

	// Create each player and start it
	for _, configPlayer := range c.Players {
		p := player.NewPlayer(configPlayer)
		go tobot.Start(p)
	}

	// block current routine
	select {}
}
