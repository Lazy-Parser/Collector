package main

import (
	"context"

	"github.com/Lazy-Parser/Collector/internal/app"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	// Starting app
	go app.Run(ctx)

	// Listening for some interruption from console (CTRL + C)
	ListenInterruptionAndStop(ctxCancel)
}
