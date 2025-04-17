package main

import (
	"context"
	"fmt"

	"Collector/internal/mexc"
)

func main() {
	fmt.Println("🚀 Starting Collector...")
	mexc.Run(context.Background())

	select {}
}
