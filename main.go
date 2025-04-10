package main

import (
	"context"
	"log"

	// "time"
	// "github.com/Lazy-Parser/Collector/internal"
	// "internal/futures"

	
	"github.com/Lazy-Parser/Collector/internal/futures"
	// "github.com/Lazy-Parser/Collector/internal/dex"
)

func main() {
	// ctx, ctxClose := context.WithTimeout(context.Background(), time.Second * 30);
	// defer ctxClose();

	
	log.Println("🚀 Starting Collector...")
	futures.Run(context.Background())

	// collector.Run(context.Background());
}
