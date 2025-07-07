package ui

import (
	"github.com/Lazy-Parser/Collector/internal/logger"
	"github.com/Lazy-Parser/Collector/internal/logic"
	"github.com/rivo/tview"
)

func InitMenuList(app *tview.Application, pages *tview.Pages) *tview.List {
	list := tview.NewList().
		AddItem("Generate Data", "Run data generation logic", '1', func() {
			pages.SwitchToPage("generate")

			logger.Get().Z.Log().Msg("I'm in Generate page!")

		}).
		AddItem("Listen Data", "Start data listener", '2', func() {
			// Call listen logic or switch to listen page

			pages.SwitchToPage("listen")
		}).
		AddItem("View Database", "View database contents", '3', func() {
			// Init table first
			pages.SwitchToPage("dbview")

			// Set table with data from database
			tokens := logic.GetDatabaseTokens()
			ui.DBView.SetTableTokens(tokens)
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			app.Stop()
		})
	list.SetBorder(true).SetTitle(" Main Menu ").SetTitleAlign(tview.AlignLeft)

	return list
}
