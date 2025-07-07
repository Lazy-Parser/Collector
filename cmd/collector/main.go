package main

import (
	"context"

	"github.com/Lazy-Parser/Collector/internal/app"
)

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	// Listening for some interruption from console (CTRL + C)
	go ListenInterruptionAndStop(ctxCancel)

	app.Init()
	if err := app.Run(ctx); err != nil {
		panic(err)
	}

	ctxCancel()
}
