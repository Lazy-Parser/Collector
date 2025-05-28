package ui

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app      *tview.Application
	tableDex *tview.Table
	tableCex *tview.Table
	flex     *tview.Flex
	logView  *tview.TextView
	paused   bool
}

var (
	ui   *UI
	once sync.Once
)

// CreateUI инициализирует UI, вызывая отдельные функции настройки
func CreateUI() {
	tableDex := createTableDex()
	tableCex := createTableCex()
	logView := newLogView()

	rootFlex := newLayout(tableDex, tableCex, logView)

	once.Do(func() {
		ui = &UI{
			app:      tview.NewApplication(),
			tableDex: tableDex,
			tableCex: tableCex,
			logView:  logView,
			flex:     rootFlex,
		}
	})

	ui.paused = false
	configureLogScrolling()
}

// createTableDex создает таблицу для DEX с необходимыми колонками
func createTableDex() *tview.Table {
	table := tview.NewTable().SetBorders(true).SetSelectable(false, false)
	table.SetTitle("DEX").SetBorder(true)
	headers := []string{"PAIR", "PRICE", "NETWORK", "POOL", "VERSION", "URL"}
	for col, h := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(fmt.Sprintf("[yellow]%s", h)).
				SetSelectable(false).
				SetAlign(tview.AlignCenter))
	}
	return table
}

// createTableCex создает таблицу для CEX с необходимыми колонками
func createTableCex() *tview.Table {
	table := tview.NewTable().SetBorders(true).SetSelectable(false, false)
	table.SetTitle("CEX").SetBorder(true)
	headers := []string{"TOKEN", "ASK", "BID"}
	for col, h := range headers {
		cell := tview.NewTableCell(fmt.Sprintf("[yellow]%s", h)).
            SetSelectable(false).
            SetAlign(tview.AlignCenter).
            SetExpansion(1)

		table.SetCell(0, col, cell)
	}
	return table
}

// newLogView создает область логов с возможностью прокрутки
func newLogView() *tview.TextView {
	logView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	logView.SetBorder(true).SetTitle("Logs (p=Pause)")
	return logView
}

// newLayout собирает основную компоновку: две таблицы сверху, лог внизу
func newLayout(dex, cex *tview.Table, logView *tview.TextView) *tview.Flex {
	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(dex, 0, 10, false).
		AddItem(cex, 0, 5, false)

	rootFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 0, 3, false).
		AddItem(logView, 0, 1, false)

	return rootFlex
}

// configureLogScrolling настраивает логику прокрутки и паузы
func configureLogScrolling() {
	ui.logView.SetChangedFunc(func() {
		if !ui.paused {
			ui.app.Draw()
		}
	})
	ui.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'p', 'P':
				ui.paused = !ui.paused
				title := "Logs"
				if ui.paused {
					title += " (PAUSED)"
				} else {
					title += " (p=Pause)"
				}
				ui.logView.SetTitle(title)
				if !ui.paused {
					ui.logView.ScrollToEnd()
				}
				return nil
			}
		}
		return event
	})
}

// GetUI возвращает синглтон UI
func GetUI() *UI { return ui }

// Run запускает приложение
func (ui *UI) Run() {
	if err := ui.app.SetRoot(ui.flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
