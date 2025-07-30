package ui

import "github.com/rivo/tview"

func InitLogBox(app *tview.Application) *tview.TextView {
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	logView.
		SetChangedFunc(func() { app.Draw(); logView.ScrollToEnd() }).
		SetBorder(true).
		SetTitle(" Logs ")

	return logView
}
