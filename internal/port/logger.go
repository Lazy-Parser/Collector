package port

import "io"

// TODO: remove SetOutput, it should not be in Logger
type Logger interface {
	Info(msg string, args ...any)
	Log(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	SetOutput(w io.Writer)
}
