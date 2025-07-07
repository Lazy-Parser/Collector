package logger

import (
	"io"
	"sync"

	"github.com/rs/zerolog"
)

var once sync.Once
var myLogger *logger

type logger struct {
	Z *zerolog.Logger
}

func New(w io.Writer) *logger {
	innerLogger := zerolog.New(w).With().Timestamp().Logger()
	l := &logger{
		Z: &innerLogger,
	}

	myLogger = l
	return l
}

func Get() *logger {
	if myLogger == nil {
		panic("logger not initialized before use")
	}
	return myLogger
}

func SetOutput(w io.Writer) {
	z := zerolog.New(w).With().Timestamp().Logger()
	myLogger.Z = &z
}
