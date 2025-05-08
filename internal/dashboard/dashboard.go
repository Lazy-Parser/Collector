package dashboard

import (
	"fmt"
	"time"

	"github.com/gosuri/uilive"
	"github.com/jedib0t/go-pretty/v6/table"
	d "github.com/Lazy-Parser/Collector/internal/domain"
)

// Run starts a live table that refreshes every `interval`.
func Run(ch <-chan d.PancakeswapV2Responce, interval time.Duration) {
	// --- live writer that rewinds cursor on every Flush()
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	// --- go-pretty table set-up (duplicate to writer)
	tw := table.NewWriter()
	tw.SetOutputMirror(writer)
	tw.AppendHeader(table.Row{"CONTRACT", "PRICE (WBNB)", "LAST SEEN"})
	tw.SetStyle(table.StyleColoredBlackOnGreenWhite)

	// ----------------------------------------------------------------------
	// cache   → latest data for each pair
	// order   → slice preserving first-seen order
	// ----------------------------------------------------------------------
	type row struct{ price float64; ts string }
	cache := map[string]row{}
	order := make([]string, 0, 32) // insertion order

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case p := <-ch:
			if _, ok := cache[p.Hex]; !ok {
				order = append(order, p.Hex) // remember first appearance
			}
			cache[p.Hex] = row{price: p.Price, ts: p.Timestamp}

		case <-ticker.C:
			tw.ResetRows()
			for _, key := range order { // stable order!
				r := cache[key]
				tw.AppendRow(table.Row{
					key,
					fmt.Sprintf("%.8f", r.price),
					r.ts,
				})
			}
			tw.Render()
			writer.Flush()
		}
	}
}
