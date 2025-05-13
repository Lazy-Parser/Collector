package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/pancakeswap_v3"
	managerDex "github.com/Lazy-Parser/Collector/internal/impl/collector/manager/dex"

	"github.com/Lazy-Parser/Collector/internal/dashboard"
	db "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/generator"

	cli "github.com/urfave/cli/v2"
)

func main() {
	db.NewConnection()

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
			{
				Name:  "db",
				Usage: "Show table of all data in database. You can ask to show pairs or tokens. By default - pairs",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "spairs",
						Aliases:  []string{"sp"},
						Usage:    "Show table of all pairs in database. By default",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "stokens",
						Aliases:  []string{"st"},
						Usage:    "Show table of all tokens in database",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "clearAll",
						Aliases:  []string{"c"},
						Usage:    "Clear all (tokens and pairs) in database",
						Required: false,
					},
				},
				Action: showTable,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runMain(*cli.Context) error {
	// init all vars

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute*3) // stop after 3 minutes
	defer ctxCancel()
	manager := managerDex.New()
	collectorDex := pancakeswap_v3.PancakeswapV3{} // pancakeswap_v2.PancakeswapV2{}
	err := manager.Push(&collectorDex)
	if err != nil {
		return fmt.Errorf("managerDex push error: %v", err)
	}

	// select pairs, that we need to update
	// pairs, err := db.GetDB().GloabalQuery(&db.Pair{Pool: "pancakeswap"}, &db.Token{Decimals: -1})
	// if err != nil {
	// 	return fmt.Errorf("fetching pairs from db: %v", err)
	// }

	// if len(pairs) != 0 { // if some tokens do not have decimals
	// 	res, err := managerDex.FetchDecimals(collectorDex.Name(), &pairs)
	// 	if err != nil {
	// 		return fmt.Errorf("managerDex errror: %v", err)
	// 	}

	// 	fmt.Printf("Fetched decimals: %d\n", len(res))
	// 	fmt.Println("Updating database...")
	// 	// update database
	// 	for address, decimal := range res {
	// 		db.GetDB().TokenService.UpdateDecimals(&db.Token{Address: address.String()}, decimal)
	// 	}

	// 	// show updates
	// 	tokens, _ := db.GetDB().TokenService.GetAllTokens()
	// 	dashboard.ShowTokens(tokens)
	// }

	// start to listen pairs

	// listen selected pairs
	// aggregator.InitJoiner()
	// joiner := aggregator.GetJoiner()

	go manager.Run(ctx)

	<-ctx.Done()

	return nil
}

func genPairs(ctx *cli.Context) error {
	generator.Run()

	// try to log all tokens from db
	fmt.Println("Generated tokens:")
	pairs, err := db.GetDB().PairService.GetAllPairs()
	if err != nil {
		return err
	}
	dashboard.ShowPairs(pairs)

	return nil
}

func showTable(ctx *cli.Context) error {
	var flag string
	if ctx.Args().Len() == 0 {
		flag = "spairs"
	}
	if ctx.Bool("spairs") {
		flag = "spairs"
	}
	if ctx.Bool("stokens") {
		flag = "stokens"
	}
	if ctx.Bool("clearAll") {
		flag = "clearAll"
	}

	if flag == "spairs" {
		// fetch pairs
		res, err := db.GetDB().PairService.GetAllPairs()
		if err != nil {
			return err
		}

		dashboard.ShowPairs(res)
	} else if flag == "stokens" {
		// fetch tokens
		res, err := db.GetDB().TokenService.GetAllTokens()
		if err != nil {
			return err
		}

		dashboard.ShowTokens(res)
	} else if flag == "clearAll" {
		err := db.GetDB().TokenService.ClearTokens()
		if err != nil {
			return err
		}

		err = db.GetDB().PairService.ClearPairs()
		if err != nil {
			return err
		}

	}

	return nil
}
