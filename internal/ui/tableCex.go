package ui

import (
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/rivo/tview"
	"strconv"
	"sync"
)

func (ui *UI) RenderTableCex(flow chan core.MexcResponse) {
	addrToRow := make(map[string]int)
	nextRow := 1

	var once sync.Once
	go func() {
		for {
			select {
			case msg := <-flow:
				if len(msg.Data.Asks) == 0 || len(msg.Data.Bids) == 0 {
					continue
				}

				// clear terminal first time
				once.Do(func() {
					fmt.Print("\033[H\033[2J")
				})

				// add if new pair
				row, exists := addrToRow[msg.Symbol]
				if !exists {
					row = nextRow
					addrToRow[msg.Symbol] = row
					nextRow++
				}

				//var average float64
				// Schedule the UI update on the tview event loop.
				ui.app.QueueUpdateDraw(func() {
					// TOKEN
					ui.tableCex.SetCell(row, 0,
						tview.NewTableCell(msg.Symbol).
							SetAlign(tview.AlignCenter))

					// ASK
					//average = findAverage(msg.Data.Asks[0][0], msg.Data.Asks[1][0], msg.Data.Asks[2][0])
					ui.tableCex.SetCell(row, 1,
						tview.NewTableCell(strconv.FormatFloat(msg.Data.Asks[0][0], 'f', 12, 64)).
							SetAlign(tview.AlignCenter))

					// BID
					//average = findAverage(msg.Data.Asks[0][0], msg.Data.Asks[1][0], msg.Data.Asks[2][0])
					ui.tableCex.SetCell(row, 2,
						tview.NewTableCell(strconv.FormatFloat(msg.Data.Bids[0][0], 'f', 12, 64)).
							SetAlign(tview.AlignCenter))
				})
			}
		}
	}()
}

//
//func findAverage(args ...float64) float64 {
//	res := 0.0
//
//	for _, arg := range args {
//		res += arg
//	}
//
//	return res / float64(len(args))
//}
