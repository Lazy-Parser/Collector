package commands

import (
	// "context"
	"fmt"
	// "time"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/dashboard"
	db "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/generator"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/solana"
	manager_dex "github.com/Lazy-Parser/Collector/internal/impl/collector/manager/dex"

	"github.com/urfave/cli/v2"
)

func Main(*cli.Context) error {
	solHelper := solana.SolanaHelper{}
	solAmmPairs, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Pool: "raydium", Label: "NULL"})

	solHelper.PushPairs(&solAmmPairs)
	manager := manager_dex.New()
	manager.FetchAndSaveMetadata(&[]core.MetadataCollector{&solHelper})

	return nil
}

func Generate(ctx *cli.Context) error {
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

func Table(ctx *cli.Context) error {
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
		// pairs, err := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "solana", Pool: "raydium"})
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

		fmt.Println("Database cleaned!")
	}

	return nil
}

// ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute*3) // stop after 3 minutes
// defer ctxCancel()
//manager := managerDex.New()
//solana := solana.Solana{}
//solanaPairs, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "solana"})
//manager.Push(&solana, &solanaPairs)
//
//solana.Init(&solanaPairs)
//list, err := solana.FetchDecimals(&solanaPairs)
//if err != nil {
//	return fmt.Errorf("failed to fetch decimals in '%s', %v", solana.Name(), err)
//}
//
//// print res
//for token, decimal := range list {
//	// save
//	db.GetDB().TokenService.UpdateDecimals(&db.Token{Address: token}, decimal)
//	fmt.Printf("Mint: %s | Decimal: %d\n", token, decimal)
//}
//fmt.Printf("Total: %d\n", len(list))
//
//fmt.Println("Saved in database!")

// manager.Push(&solana, &solanaPairs)

// err := manager.Push(&pancakeswapV2, &psV2pairs)
// if err != nil {
// 	return fmt.Errorf("managerDex push error: %v", err)
// }
// err = manager.Push(&pancakeswapV3, &psV3pairs)
// if err != nil {
// 	return fmt.Errorf("managerDex push error: %v", err)
// }

// go manager.Run(ctx)

// <-ctx.Done()

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
