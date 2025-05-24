package dashboard

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"strconv"

	database "github.com/Lazy-Parser/Collector/internal/database"
)

func ShowPairs(pairs []database.Pair) {
	// --- go-pretty table set-up (duplicate to writer)
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(table.Row{"ID", "PAIR", "Base ID", "Quote ID", "PAIR CONTRACT", "NETWORK", "POOL", "LABEL"})
	tw.SetStyle(table.StyleColoredBlackOnGreenWhite)

	var rows []table.Row
	for _, pair := range pairs {
		row := table.Row{
			pair.ID,
			pair.BaseToken.Name + "/" + pair.QuoteToken.Name,
			pair.BaseTokenID,
			pair.QuoteTokenID,
			pair.PairAddress,
			pair.Network,
			pair.Pool,
			pair.Label,
		}
		rows = append(rows, row)
	}

	tw.AppendRows(rows)
	tw.AppendSeparator()
	tw.AppendFooter(table.Row{"", "", "", "", "", "Total", len(pairs), ""})

	fmt.Println("PAIRS")
	tw.Render()
}

func ShowTokens(tokens []database.Token) {
	// --- go-pretty table set-up (duplicate to writer)
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(table.Row{"ID", "NAME", "ADDRESS", "DECIMALS", "VAULTS (SOLANA)"})
	tw.SetStyle(table.StyleColoredBlackOnGreenWhite)

	var rows []table.Row
	for _, token := range tokens {
		row := table.Row{
			token.ID,
			token.Name,
			token.Address,
			showDecimals(token.Decimals),
			showVaults(token.Vault),
		}
		rows = append(rows, row)
	}

	tw.AppendRows(rows)
	tw.AppendSeparator()
	tw.AppendFooter(table.Row{"", "", "", "Total", len(tokens)})

	fmt.Println("TOKENS")
	tw.Render()
}

func showDecimals(d int) string {
	if d == -1 {
		return "NULL"
	}

	return strconv.Itoa(d)
}

func showVaults(str string) string {
	if str == "" {
		return "NULL"
	}

	return str
}
