package main

import (
	"log"
	"os"

	"github.com/Lazy-Parser/Collector/internal/generator"
	"github.com/Lazy-Parser/Collector/internal/impl/aggregator"
	mexc "github.com/Lazy-Parser/Collector/internal/impl/collector/cex"
	m "github.com/Lazy-Parser/Collector/internal/impl/collector/manager"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "collector",
		Usage: "Service, that collects pairs prices and publish to NATS",
		Commands: []*cli.Command{
			{
				Name:   "collect",
				Usage:  "DEX/CEX data collector. DEX listen only selected pairs. You can generate those pairs by using 'generate' command",
				Action: runMain,
			},
			{
				Name:  "generate",
				Usage: "Generate pairs and save to JSON. Get all pairs from CEX, filter by volume, fetch addresses, networks name, liquidity pool name, ...",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "volume",
						Aliases:  []string{"e"},
						Usage:    "Write to JSON only pairs, that > 'volume'. 1,000,000 by default",
						Required: false,
					},
				},
				Action: genPairs,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runMain(*cli.Context) error {
	aggregator.InitJoiner()
	joiner := aggregator.GetJoiner()

	manager := m.CreateManager()
	manager.NewCollector(&mexc.MexcSource{})

	go manager.Run(joiner)

	return nil
}

func genPairs(ctx *cli.Context) error {

	// generator.Run(&mexc.MexcSource{})
	generator.Run()

	return nil
}
