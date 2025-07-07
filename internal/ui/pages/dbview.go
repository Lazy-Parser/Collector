package pages

import (
	"strconv"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var defaultText = "Database Viewer Page. Press Esc to return to menu"

type DBView struct {
	Flex        *tview.Flex
	Text        *tview.TextView
	TableTokens *tview.Table
	TablePairs  *tview.Table
}

// TODO: add token and pair viewer. Just create one more flex for tables and add one more table to the right
func InitPageDBView(pages *tview.Pages) *DBView {
	// Text
	text := tview.NewTextView()
	text.SetText(defaultText)

	// TABLES
	//
	// Table tokens
	tableTokens := tview.NewTable().SetBorders(true)
	tableTokens.SetTitle("TOKENS")
	headers := []string{"NAME", "DECIMALS", "ADDRESS"}
	for i, h := range headers {
		tableTokens.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	// Table pairs
	tablePairs := tview.NewTable().SetBorders(true)
	tablePairs.SetTitle("PAIRS")
	headers = []string{"ID", "PAIR", "Base ID", "Quote ID", "PAIR CONTRACT", "NETWORK", "POOL", "LABEL", "TYPE"}
	for i, h := range headers {
		tablePairs.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	// FLEX
	//
	// Flex Tables
	flexTables := tview.NewFlex()
	flexTables.SetBorder(false)
	flexTables.SetDirection(tview.FlexColumn).AddItem(tablePairs, 0, 2, false).AddItem(tableTokens, 0, 1, false)

	// Flex global
	flex := tview.NewFlex()
	flex.SetBorder(true).
		SetTitle(" Database ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				text.SetText(defaultText)

				pages.SwitchToPage("menu")
				return nil
			}
			return event
		})
	flex.SetDirection(tview.FlexRow).AddItem(text, 0, 1, false).AddItem(flexTables, 0, 10, true)

	// create class
	dbview := DBView{
		Flex:        flex,
		Text:        text,
		TableTokens: tableTokens,
		TablePairs:  tablePairs,
	}

	return &dbview
}

func (dbview *DBView) SetTableTokens(tokens []core.Token) {
	for i, token := range tokens {
		i += 1

		dbview.TableTokens.SetCell(i, 0, tview.NewTableCell(token.Name).SetAlign(tview.AlignCenter))
		dbview.TableTokens.SetCell(i, 1, tview.NewTableCell(strconv.FormatUint(uint64(token.Decimal), 10)).SetAlign(tview.AlignCenter))
		dbview.TableTokens.SetCell(i, 2, tview.NewTableCell(token.Address).SetAlign(tview.AlignCenter))
	}
}

func (dbview *DBView) SetTablePairs(tokens []core.Token) {

	// for i, token := range tokens {
	// 	i += 1

	// }
}
