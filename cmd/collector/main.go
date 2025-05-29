package main

import (
	"log"
	"os"

	"github.com/Lazy-Parser/Collector/cmd/collector/commands"
	db "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/urfave/cli/v2"
)

func main() {
	db.NewConnection()

	app := &cli.App{
		Name:     "collector",
		Usage:    "Service, that collects pairs prices and publish to NATS",
		Commands: commands.MyCommands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
