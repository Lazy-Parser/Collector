package ui

import (
	"sync"

	"github.com/rivo/tview"
)

var (
	ui   *UI
	once sync.Once
)

type UI struct {
	app  *tview.Application
	list *tview.List
}

func Create() *UI {
	app := tview.NewApplication()

	// MENU LIST
	list := tview.NewList().
		AddItem("Generate Data", "Run data generation logic", '1', func() {
			// Call your generate function or switch to generate page
			// Example:
			// svc.GenerateData()
		}).
		AddItem("Listen Data", "Start data listener", '2', func() {
			// Call listen logic or switch to listen page
		}).
		AddItem("View Database", "View database contents", '3', func() {
			// Call db view logic or switch to db page
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			app.Stop()
		})
	list.SetBorder(true).SetTitle(" Main Menu ").SetTitleAlign(tview.AlignLeft)

	userInterface := UI{}
	userInterface.app = app
	userInterface.list = list
	ui = &userInterface

	return &userInterface
}

func GetUI() *UI {
	return ui
}

func (ui *UI) Run() {
	if err := ui.app.SetRoot(ui.list, true).Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) Stop() {
	ui.app.QueueUpdateDraw(func() {
		ui.app.Stop()
	})
}
