package main

import (
	"os"
	"path/filepath"

	config "github.com/Lazy-Parser/Collector/config/service"
	"github.com/Lazy-Parser/Collector/internal/app"
)

func main() {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "..", ".env")

	cfg, err := config.NewConfig(path)
	if err != nil {
		panic(err)
	}

	if err := app.Run(cfg); err != nil {
		panic(err)
	}
}
