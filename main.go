package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
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
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load(*configFile)
	if err != nil {
		slog.Error("cannot load config", "err", err)
		os.Exit(1)
	}

	variant := variants.Get(cfg.Variant)
	if variant == nil {
		slog.Error("this variant does not exists", "variant", cfg.Variant)
	}

	game := game.Init(cfg.Game, variant)
	server := server.New(cfg.Server, game, variant)

	server.Start()
}
