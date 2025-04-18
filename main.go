package main

import (
	"context"
	"fmt"
	"os"

	"Collector/internal/mexc"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🚀 Starting Collector...")
	if err := godotenv.Load(); err != nil {
		fmt.Println("sdsdfsdfsdfsdf")
	}

	mexcFutures := os.Getenv("MEXC_FUTURES_WS")
	mexcSpot := os.Getenv("MEXC_SPOT_WS")

	fmt.Println("mexcFutures")
	fmt.Println(mexcFutures)
	fmt.Println("mexcSpot")
	fmt.Println(mexcSpot)
	mexc.Run(context.Background())
}
