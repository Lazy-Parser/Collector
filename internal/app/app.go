package app

import (
	"context"

	"github.com/Lazy-Parser/Collector/internal/ui"
)

func initCleopatra(ctx context.Context) {
	ui.Create()

	// ...
}

func Run(ctx context.Context) {
	initCleopatra(ctx)

	go ui.GetUI().Run()
	// ...

	// Stopping all services by ctx signal
	<-ctx.Done()
	ui.GetUI().Stop()
}
