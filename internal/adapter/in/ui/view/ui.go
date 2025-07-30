package ui

import (
	"context"
	"sync"

	config "Cleopatra/config/service"
	p "Cleopatra/internal/adapter/in/ui/view/pages"
	generator "Cleopatra/internal/generator/usecase"
	"Cleopatra/internal/port"

	"github.com/rivo/tview"
)

var (
	ui   *UI
	once sync.Once
)

type UI struct {
	app      *tview.Application
	textView *tview.TextView
	layout   *tview.Flex
	pages    *tview.Pages

	DBView *p.DBView
}

type GeneratorRepo interface {
	GenerateMexc(ctx context.Context, cfg *config.Config, progress chan generator.Progress) error
}

func Create(logger port.Logger, generator GeneratorRepo) *UI {
	app := tview.NewApplication()

	// LOG BOX
	logBox := InitLogBox(app)

	// LOGGER set ui's writer
	writer := NewTviewWriter(app, logBox)
	logger.SetOutput(writer)

	// PAGES
	pages := tview.NewPages()
	generate := p.InitPageGenerate(pages)
	listen := p.InitPageListen(pages)
	dbView := p.InitPageDBView(pages, app)

	pages.AddPage("generate", generate, true, false)
	pages.AddPage("listen", listen, true, false)
	pages.AddPage("dbview", dbView.Flex, true, false)

	menu := InitMenuList(app, pages, logger, generator)
	pages.AddPage("menu", menu, true, true)

	// LAYOUT
	layout := tview.NewFlex().
		AddItem(pages, 0, 3, true).  // left: 3/4 width
		AddItem(logBox, 0, 1, false) // right: 1/4 width

	userInterface := UI{}
	userInterface.app = app
	userInterface.layout = layout
	userInterface.textView = logBox
	userInterface.DBView = dbView
	userInterface.pages = pages
	ui = &userInterface

	return &userInterface
}

func GetUI() *UI {
	return ui
}

func (ui *UI) GetApp() *tview.Application {
	return ui.app
}

func (ui *UI) GetLogBox() *tview.TextView {
	return ui.textView
}

func (ui *UI) Run() error {
	return ui.app.SetRoot(ui.layout, true).Run()
}

func (ui *UI) Stop() {
	ui.app.QueueUpdateDraw(func() {
		ui.app.Stop()
	})
}
