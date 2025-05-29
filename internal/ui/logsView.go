package ui

import "fmt"

func (ui *UI) LogsView(logLine string) {
	ui.app.QueueUpdateDraw(func() {
		fmt.Fprintf(ui.logView, "[white]%s\n", logLine)
		if !ui.paused {
			ui.logView.ScrollToEnd()
		}
	})
}
