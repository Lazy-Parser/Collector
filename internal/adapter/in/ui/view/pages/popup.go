package pages

import (
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

	pages.AddPage("popup", modal, true, true)
}
