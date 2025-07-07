package pages

import (
	"fmt"
	"strconv"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

var defaultText = "Database Viewer Page. Press Esc to return to menu.\nPress TAB to change table's focus.\nPress Q / W to set focus to pair table / token table."

type DBView struct {
	Flex        *tview.Flex
	Text        *tview.TextView
	TableTokens *tview.Table
	TablePairs  *tview.Table
}

// TODO: a lot of logic in one func, make it all separate
func InitPageDBView(pages *tview.Pages, app *tview.Application) *DBView {
	// Text
	text := tview.NewTextView()
	text.SetText(defaultText)

	// TABLES
	//
	// Table tokens
	tableTokens := tview.NewTable().SetBorders(true).SetSelectable(true, false)
	tableTokens.SetTitle("TOKENS")
	headers := []string{"NAME", "DECIMALS", "ADDRESS"}
	for i, h := range headers {
		tableTokens.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}
	tableTokens.SetSelectedFunc(func(row int, column int) {
		name := tableTokens.GetCell(row, 0).Text
		decimal := tableTokens.GetCell(row, 1).Text
		address := tableTokens.GetCell(row, 2).Text

		msg := fmt.Sprintf(
			"Name: %s\n Decimal: %s \nAddress: %s",
			name, decimal, address,
		)

		oldFocus := app.GetFocus()
		ShowPopup(
			pages,
			msg,
			func() {
				app.SetFocus(oldFocus)
				clipboard.Write(clipboard.FmtText, []byte(address))
			},
		)
	})

	// Table pairs
	tablePairs := tview.NewTable().SetBorders(true).SetSelectable(true, false)
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
	flexTables.SetDirection(tview.FlexColumn).AddItem(tablePairs, 0, 2, true).AddItem(tableTokens, 0, 1, true)

	// Flex global
	flex := tview.NewFlex()
	flex.SetBorder(true).
		SetTitle(" Database ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// If ESC pressed - exit from current page
			if event.Key() == tcell.KeyEsc {
				text.SetText(defaultText)

				pages.SwitchToPage("menu")
				return nil
			}

			// TAB - change tables focus
			if event.Key() == tcell.KeyTAB {
				if tablePairs.HasFocus() {
					setFocus(app, tableTokens, text)
				} else {
					setFocus(app, tablePairs, text)
				}

				return nil
			}

			// Q - change focus to the pair table
			// W - change focus to the token table
			if event.Key() == tcell.KeyRune {
				switch event.Rune() {
				case 'q':
					setFocus(app, tablePairs, text)
				case 'w':
					setFocus(app, tableTokens, text)
				}

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
	// Insert new data
	for i, token := range tokens {
		rowIndex := i + 1 // skip header

		dbview.TableTokens.SetCell(rowIndex, 0,
			tview.NewTableCell(token.Name).
				SetAlign(tview.AlignCenter).
				SetSelectable(true))
		dbview.TableTokens.SetCell(rowIndex, 1,
			tview.NewTableCell(strconv.FormatUint(uint64(token.Decimal), 10)).
				SetAlign(tview.AlignCenter).
				SetSelectable(true))
		dbview.TableTokens.SetCell(rowIndex, 2,
			tview.NewTableCell(token.Address).
				SetAlign(tview.AlignCenter).
				SetSelectable(true))
	}

	// Select first data row to enable navigation
	if len(tokens) > 0 {
		dbview.TableTokens.Select(1, 0)
	}
}

func (dbview *DBView) SetTablePairs(tokens []core.Token) {

	// for i, token := range tokens {
	// 	i += 1

	// }
}

func setFocus(app *tview.Application, p *tview.Table, text *tview.TextView) {
	app.SetFocus(p)
	msg := fmt.Sprintf("\nCurrent focus: %s", p.GetTitle())
	text.SetText(defaultText + msg)
}
