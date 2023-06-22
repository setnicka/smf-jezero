package main

import (
	"flag"
	"log"

	"github.com/setnicka/smf-jezero/config"
	"github.com/setnicka/smf-jezero/game"
	"github.com/setnicka/smf-jezero/game/variants"
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

	variant := variants.Get(cfg.Variant)
	if variant == nil {
		log.Fatalf("Variant '%s' does not exist", cfg.Variant)
	}

	game := game.Init(cfg.Game, variant)
	server := server.New(cfg.Server, game, variant)

	server.Start()
}
