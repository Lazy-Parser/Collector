package ui

import "fmt"

// t (type) - "error" | "log" | "warning"
func (ui *UI) LogsView(logLine string, t string) {
	ui.App.QueueUpdateDraw(func() {
		if t == "error" {
			fmt.Fprintf(ui.logView, "[red]%s\n", logLine)
		} else if t == "log" {
			fmt.Fprintf(ui.logView, "[white]%s\n", logLine)
		} else if t == "warning" {
			fmt.Fprintf(ui.logView, "[yellow]%s\n", logLine)
		}

		if !ui.paused {
			ui.logView.ScrollToEnd()
		}
	})
}
