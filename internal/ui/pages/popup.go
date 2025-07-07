package pages

import (
	"github.com/Lazy-Parser/Collector/internal/logger"
	"github.com/rivo/tview"
)

// ShowPopup - create a small info window.
// cb - is a callback function, that will be called when closing popup
func ShowPopup(pages *tview.Pages, text string, cb func()) {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("popup")
			cb()
		})
	modal.SetTitle(" Info ").SetBorder(true)

	logger.Get().Z.Info().Msg("POPUS")

	pages.AddPage("popup", modal, true, true)
}
