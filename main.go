package main

import (
	"context"
	"fmt"

	"Collector/internal/mexc"
)

func main() {
	fmt.Println("🚀 Starting Collector...")
	test()
	mexc.Run(context.Background())
	test()
}

func test() {
	fmt.Println("TEST")
}
