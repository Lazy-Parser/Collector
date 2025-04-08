package main

import (
	"log"
	"context"
	"github.com/Lazy-Parser/Collector/internal"
)

func main() {
	log.Println("🚀 Starting Collector...")
	collector.Run(context.Background());
}