package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func InitPageListen(pages *tview.Pages) *tview.TextView {
	listen := tview.NewTextView()
	listen.SetText("Listen Data Page\nPress Esc to return to menu").
		SetBorder(true).SetTitle(" Listen ")

	listen.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {

			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	return listen
}
