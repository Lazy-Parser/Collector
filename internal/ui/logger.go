package ui

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/rivo/tview"
)

type tviewWriter struct {
	app      *tview.Application
	textView *tview.TextView
	mu       sync.Mutex
}

// NewTviewWriter implements Write method from io.Writer. It should be use to pass into Logger to show logs in UI in LogBox
func NewTviewWriter(app *tview.Application, textView *tview.TextView) *tviewWriter {
	return &tviewWriter{
		app:      app,
		textView: textView,
	}
}

func (tw *tviewWriter) Write(p []byte) (n int, err error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	var entry struct {
		Level   string    `json:"level"`
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}
	if err := json.Unmarshal(p, &entry); err != nil {
		// If parsing fails, write raw line
		go func() {
			tw.app.QueueUpdate(func() {
				tw.textView.Write(p)
				tw.textView.ScrollToEnd()
			})
		}()

		return len(p), nil
	}

	var prefix string
	switch entry.Level {
	case "info":
		prefix = "[blue]"
	case "warn":
		prefix = "[orange]"
	case "error":
		prefix = "[red]"
	default: // if 'level' value is empty, its 'log'
		prefix = "[white]"
	}

	str := prefix + entry.Message + "\n"
	// push to ui logger
	go func() {
		tw.app.QueueUpdateDraw(func() {
			tw.textView.Write([]byte(str))
			tw.textView.ScrollToEnd()
		})
	}()

	return len(p), nil
}
