package app

import (
	"context"
	"os"

	"github.com/Lazy-Parser/Collector/internal/logger"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"golang.design/x/clipboard"
)

func Init() error {
	// Init clipboard
	err := clipboard.Init()
	if err != nil {
		return err
	}

	// Create logger with default Writer
	logger.New(os.Stdout)

	// UI will change logger's Writer to custom while ui creation
	ui.Create()

	// ...

	return nil
}

func Run(ctx context.Context) error {

	logger.Get().Z.Info().Msg("Program started!")
	return ui.GetUI().Run()
}
