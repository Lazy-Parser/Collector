package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func InitPageGenerate(pages *tview.Pages) *tview.TextView {
	generate := tview.NewTextView()
	generate.SetText("Generate Data Page\nPress Esc to return to menu").
		SetBorder(true).SetTitle(" Generate ")

	generate.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("menu")

			// logger.Get().Z.Warn().Msg("Exiting from Generate Page to menu!")

			return nil
		}
		return event
	})

	return generate
}
