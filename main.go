package main

import (
	"context"
	"fmt"

	"github.com/Lazy-Parser/Collector/internal/mexc"
)

func main() {
	fmt.Println("🚀 Starting Collector...")
	mexc.Run(context.Background())
}
