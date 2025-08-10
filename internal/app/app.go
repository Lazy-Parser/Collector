package app

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	database "github.com/Lazy-Parser/Collector/internal/adapter/out/persistent/sqlite"
	api_coingecko "github.com/Lazy-Parser/Collector/internal/api/coingecko"
	api_dexscreener "github.com/Lazy-Parser/Collector/internal/api/dexscreener"
	api_mexc "github.com/Lazy-Parser/Collector/internal/api/mexc"
	"github.com/Lazy-Parser/Collector/internal/application/generator"
	"github.com/Lazy-Parser/Collector/internal/common/chains"
	logger "github.com/Lazy-Parser/Collector/internal/common/zerolog"
	worker_coingecko "github.com/Lazy-Parser/Collector/internal/domain/api/coingecko"
	worker_dexscreener "github.com/Lazy-Parser/Collector/internal/domain/api/dexscreener"
	worker_mexc "github.com/Lazy-Parser/Collector/internal/domain/api/mexc"
	"context"
	"os"
	"path/filepath"
)

func Run(cfg *config.Config) error {
	generatorService, db, err := Init(cfg)
	if err != nil {
		return err
	}

	return RunGenerator(generatorService, db)
}

func Init(cfg *config.Config) (*generator.GeneratorService, *database.Database, error) {
	logger.New(os.Stdout)

	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(wd, "..", "..", "config", "configs", "chains.json")
	chainsService, err := chains.NewChains(path)
	if err != nil {
		return nil, nil, err
	}

	mexcApi := api_mexc.NewMexcAPI()
	mexcWorker := worker_mexc.NewWorker(mexcApi, cfg, chainsService)

	dsApi := api_dexscreener.NewDexscreenerAPI(cfg)
	dsWorker := worker_dexscreener.NewWorker(dsApi, chainsService)

	cgApi := api_coingecko.NewCoingeckoAPI(cfg)
	cgWorker := worker_coingecko.NewWorker(cfg, cgApi)

	gen := generator.NewGeneratorService(mexcWorker, dsWorker, cgWorker)

	dbPath := filepath.Join(wd, "..", "..", "storage", "collector.sqlite")
	db, err := database.NewConnection(dbPath)
	if err != nil {
		return nil, nil, err
	}
	return gen, db, nil
}

func RunGenerator(service *generator.GeneratorService, db *database.Database) error {
	ctx := context.Background()

	// ----- FUTURES --------
	futures, err := service.GetFutures(ctx)
	if err != nil {
		return err
	}

	for _, token := range futures {
		if err := db.SaveToken(token); err != nil {
			logger.Get().Z.Error().Msgf("Failed to save token into database: %v", err)
		}
	}
	// ----- FUTURES --------

	// ----- PAIRS --------
	// pairs, err := service.GetPairs(ctx, futures)
	// if err != nil {
	// 	return fmt.Errorf("failed to get pairs: %v", err)
	// }
	// for _, pair := range pairs {
	// 	if err := db.SavePair(pair); err != nil {
	// 		logger.Get().Z.Error().Msgf("Failed to save pair into database: %v", err)
	// 	}
	// }
	// ----- PAIRS --------

	return nil
}
