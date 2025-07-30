package ui

import (
	config "Cleopatra/config/service"
	"Cleopatra/internal/adapter/in/ui/controller"
	"Cleopatra/internal/port"
	"context"

	"github.com/rivo/tview"
)

func InitMenuList(app *tview.Application, pages *tview.Pages, logger port.Logger, generator GeneratorRepo) *tview.List {
	list := tview.NewList().
		AddItem("Generate Data", "Run data generation logic", '1', func() {
			pages.SwitchToPage("generate")

			params := &controller.Params{
				Logger:    logger,
				Generator: generator,
			}

			generatorController := controller.NewGeneratorController(params)
			generatorController.TryGenerate(context.Background(), &config.Config{})

			// logger.Get().Z.Log().Msg("I'm in Generate page!")

		}).
		AddItem("Listen Data", "Start data listener", '2', func() {
			// Call listen logic or switch to listen page

			// Move to the page first
			pages.SwitchToPage("listen")

			// Init data transmission
			//..

			// Init listener and pass data transmission

		}).
		AddItem("View Database", "View database contents", '3', func() {
			// Init table first
			pages.SwitchToPage("dbview")

			// // Set table with data from database
			// tokens, err := tokenPairService.GetAllTokens()
			// if err != nil {
			// 	logger.Error("Level: %s .Failed to get all tokens from database: %v", err)
			// }

			logger.Info("I'm in Viewer page!")

			// ui.DBView.SetTableTokens(tokens)
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			app.Stop()
		})
	list.SetBorder(true).SetTitle(" Main Menu ").SetTitleAlign(tview.AlignLeft)

	return list
}
