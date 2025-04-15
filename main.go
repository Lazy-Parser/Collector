package main

import (
	"context"
	"log"

	"github.com/Lazy-Parser/Collector/internal/mexc"
)

func main() {
	log.Println("🚀 Starting Collector...")
	mexc.Run(context.Background())
}
