package logger

import (
	"Cleopatra/internal/port"
	"io"

	"github.com/rs/zerolog"
)

type Logger struct {
	Z *zerolog.Logger
}

// compile‑time гарантия
var _ port.Logger = (*Logger)(nil)

func New(w io.Writer) *Logger {
	innerLogger := zerolog.New(w).With().Timestamp().Logger()
	return &Logger{
		Z: &innerLogger,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	z := zerolog.New(w).With().Timestamp().Logger()
	l.Z = &z
}

func (l *Logger) Info(msg string, args ...any) {
	l.Z.Info().Msgf(msg, args...)
}

func (l *Logger) Log(msg string, args ...any) {
	l.Z.Log().Msgf(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Z.Warn().Msgf(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Z.Error().Msgf(msg, args...)
}
