package main

import (
	// "context"
	// "encoding/json"
	// "fmt"
	// "fmt"
	"fmt"
	"log"
	"os"

	// "path/filepath"

	db "github.com/Lazy-Parser/Collector/internal/database"
	manager_dex "github.com/Lazy-Parser/Collector/internal/impl/collector/manager/dex"

	// "github.com/Lazy-Parser/Collector/internal/domain"

	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/pancakeswap_v2"

	// "github.com/ethereum/go-ethereum/common"

	// "github.com/ethereum/go-ethereum/common"
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
	// aggregator.InitJoiner()
	// joiner := aggregator.GetJoiner()

	// manager := m.CreateManager()
	// manager.NewCollector(&mexc.MexcSource{})

	// go manager.Run(joiner)

	db.NewConnection()
	managerDex := manager_dex.New()
	collectorDex := pancakeswap_v2.PancakeswapV2{}

	managerDex.Push(collectorDex)

	return nil
}

func genPairs(ctx *cli.Context) error {

	db.NewConnection()
	// db.GetDB().ClearTokens()
	// db.GetDB().ClearPairs()
	// generator.Run()

	// try to log all tokens from db
	var res []db.Pair
	res, err := db.GetDB().GetAllPairs()
	if err != nil {
		return err
	}

	for idx, pair := range res {
		fmt.Printf(
			"%d) %s/%s. Network: %s | Pool: %s\n",
			idx, pair.BaseToken.Name, pair.QuoteToken.Name,
			pair.Network, pair.Pool,
		)
	}

	// pair.BaseToken.Name, pair.QuoteToken.Name не отображаеться, починить

	return nil
}
