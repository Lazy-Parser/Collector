package commands

import "github.com/urfave/cli/v2"

var MyCommands = []*cli.Command{
	{
		Name:   "collect",
		Usage:  "DEX/CEX data collector. DEX listen only selected pairs. You can generate those pairs by using 'generate' command",
		Action: Main,
	},
	{
		Name:  "generate",
		Usage: "Generate pairs and save to JSON. Get all pairs from CEX, filter by volume, fetch addresses, networks name, liquidity pool name, ...",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "volume",
				Aliases:  []string{"e"},
				Usage:    "Write to JSON only pairs, that > 'volume'. 1,000,000 by default",
				Required: false,
			},
		},
		Action: Generate,
	},
	{
		Name:  "db",
		Usage: "Show table of all data in database. You can ask to show pairs or tokens. By default - pairs",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "spairs",
				Aliases:  []string{"sp"},
				Usage:    "Show table of all pairs in database. By default",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "stokens",
				Aliases:  []string{"st"},
				Usage:    "Show table of all tokens in database",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "clearAll",
				Aliases:  []string{"c"},
				Usage:    "Clear all (tokens and pairs) in database",
				Required: false,
			},
		},
		Action: Table,
	},
}
