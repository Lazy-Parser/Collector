package ui

import (
	"fmt"
	"github.com/rivo/tview"
	"sync"
)

type UI struct {
	app     *tview.Application
	table   *tview.Table
	flex    *tview.Flex
	logView *tview.TextView
}

var (
	ui   *UI
	once sync.Once
)

func CreateUI() {
	once.Do(func() {
		ui = &UI{}
		ui.app = tview.NewApplication() // init
		ui.table = tview.NewTable().SetBorders(true).SetSelectable(false, false)
		ui.logView = tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() { ui.app.Draw() })
		ui.flex = tview.NewFlex().
			AddItem(ui.table, 0, 3, false).
			AddItem(ui.logView, 0, 2, false)
	})
	ui.logView.SetBorder(true).SetTitle("Logs")

	// Set up the header row.
	headers := []string{"PAIR", "PRICE", "NETWORK", "POOL", "VERSION", "URL", "ISBASETOKEN0"}
	for col, h := range headers {
		ui.table.SetCell(0, col,
			tview.NewTableCell(fmt.Sprintf("[yellow]%s", h)).
				SetSelectable(false).
				SetAlign(tview.AlignCenter))
	}
}

func GetUI() *UI { return ui }

// call in separete goroutine
func (ui *UI) Run() {
	if err := ui.app.SetRoot(ui.flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
