package main

import (
	"flag"
	"log"

	"github.com/setnicka/smf-jezero/config"
	"github.com/setnicka/smf-jezero/game"
	"github.com/setnicka/smf-jezero/server"
)

var (
	configFile = flag.String("config", "configuration.json", "Configuration file")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	game := game.Init()
	server := server.New(cfg.Server, game)

	server.Start()
}
