package main

import (
	"context"
	"fmt"

	"Collector/internal/mexc"
)

func main() {
	fmt.Println("🚀 Starting Collector...")
	go mexc.Run(context.Background())
}
