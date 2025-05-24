package ui

import (
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/rivo/tview"
	"strconv"
	"sync"
)

func (ui *UI) ShowCollectorPrices(flow chan core.CollectorDexResponse) {

	allPairs, _ := database.GetDB().PairService.GetAllPairs()

	addrToRow := make(map[string]int)
	nextRow := 1

	var once sync.Once
	go func() {
		for {
			select {
			case msg := <-flow:
				// clear terminal first time
				once.Do(func() {
					fmt.Print("\033[H\033[2J")
				})

				// add if new pair
				row, exists := addrToRow[msg.Address]
				if !exists {
					row = nextRow
					addrToRow[msg.Address] = row
					nextRow++
				}

				// find pair
				var pair *database.Pair
				for _, p := range allPairs {
					if p.PairAddress == msg.Address {
						pair = &p
						break
					}
				}
				if pair == nil {
					fmt.Printf("Cannot find pair by address '%s' in dashboard!", msg.Address)
					continue
				}

				// Schedule the UI update on the tview event loop.
				ui.app.QueueUpdateDraw(func() {
					// PAIR
					ui.table.SetCell(row, 0,
						tview.NewTableCell(pair.BaseToken.Name+"/"+pair.QuoteToken.Name).
							SetAlign(tview.AlignCenter))

					// PRICE
					ui.table.SetCell(row, 1,
						tview.NewTableCell(msg.Price.Text('f', 12)).
							SetAlign(tview.AlignCenter))

					// NETWORK
					ui.table.SetCell(row, 2,
						tview.NewTableCell(pair.Network).
							SetAlign(tview.AlignCenter))

					// POOL
					ui.table.SetCell(row, 3,
						tview.NewTableCell(pair.Pool).
							SetAlign(tview.AlignCenter))

					// VERSION
					ui.table.SetCell(row, 4,
						tview.NewTableCell(pair.Label).
							SetAlign(tview.AlignCenter))

					// URL
					ui.table.SetCell(row, 5,
						tview.NewTableCell(pair.URL).
							SetAlign(tview.AlignRight))

					// IsBaseToken0
					ui.table.SetCell(row, 6,
						tview.NewTableCell(strconv.FormatBool(msg.IsBaseToken0)).
							SetAlign(tview.AlignRight))
				})
			}
		}
	}()
}
