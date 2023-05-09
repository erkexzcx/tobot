package main

import (
	"flag"
	"os"
	"tobot"

	"tobot/comms"
	"tobot/config"
	"tobot/player"

	"github.com/op/go-logging"

	_ "tobot/module/all"
)

var configPath = flag.String("config", "config.yml", "path to config file")
var log = logging.MustGetLogger("global")

func main() {
	// Parse command line arguments
	flag.Parse()

	// Parse configuration file
	c, err := config.NewConfig(*configPath)
	if err != nil {
		log.Panic("Configuration failed:", err)
	}

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	loggerFormat := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} [%{module}] %{message}`)
	logbackendFormatter := logging.NewBackendFormatter(logBackend, loggerFormat)
	logbackendLeveled := logging.AddModuleLevel(logbackendFormatter)
	switch c.LogLevel {
	case "DEBUG":
		logbackendLeveled.SetLevel(logging.DEBUG, "")
	case "INFO":
		logbackendLeveled.SetLevel(logging.INFO, "")
	case "WARNING":
		logbackendLeveled.SetLevel(logging.WARNING, "")
	case "CRITICAL":
		logbackendLeveled.SetLevel(logging.CRITICAL, "")
	}
	logging.SetBackend(logbackendFormatter)
	log.Debug("Logger initialized")

	// Start comms package
	comms.InitComms(c)
	log.Debug("Comms initialized")

	// Create each player and start it
	for _, configPlayer := range c.Players {
		p := player.NewPlayer(configPlayer)
		log.Debug("Player created:", p.Config.Nick)
		go tobot.Start(p)
	}

	// block current routine
	select {}
}
