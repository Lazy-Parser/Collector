package app

import (
	"context"
	"os"

	"github.com/Lazy-Parser/Collector/internal/logger"
	"github.com/Lazy-Parser/Collector/internal/ui"
)

func Init() {
	// Create logger with default Writer
	logger.New(os.Stdout)

	// UI will change logger's Writer to custom while ui creation
	ui.Create()

	// ...
}

func Run(ctx context.Context) error {

	logger.Get().Z.Info().Msg("Program started!")
	return ui.GetUI().Run()
}
