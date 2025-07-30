package app

import (
	config "Cleopatra/config/service"
	"fmt"
	"os"
	"path/filepath"

	ui_main "Cleopatra/internal/adapter/in/ui"
	logger "Cleopatra/internal/adapter/out/log/zerolog"
	database "Cleopatra/internal/adapter/out/persistent/sqlite"
	"Cleopatra/internal/adapter/out/webapi/mexc"
	generator "Cleopatra/internal/generator/usecase"
	market_usecase "Cleopatra/internal/market/usecase"
)

func Run(cfg *config.Config) error {
	logger := logger.New(os.Stdout)

	wd, _ := os.Getwd()
	dbPath := filepath.Join(wd, "storage", "collector.db")
	db, err := database.NewConnection(dbPath)
	if err != nil {
		return err
	}

	tokenService := market_usecase.NewTokenService(db, logger)
	pairService := market_usecase.NewTokenService(db, logger)

	mexc := mexc.NewMexc()
	generatorService := generator.NewGenerator(logger, db, mexc)

	params := &ui_main.Params{
		Logger:       logger,
		TokenService: tokenService,
		PairService:  pairService,
		Generator:    generatorService,
	}
	if err := ui_main.Run(params); err != nil {
		return fmt.Errorf("app: %v", err)
	}

	return nil
}
