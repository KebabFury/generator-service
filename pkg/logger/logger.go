package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var consoleWriter io.Writer = zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: time.RFC3339,
}

func NewConsole() zerolog.Logger {
	return zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger()
}

var once sync.Once

var log zerolog.Logger

func Get() zerolog.Logger {
	once.Do(func() {
		log = zerolog.New(consoleWriter).
			With().
			Timestamp().
			Logger()
	})

	return log
}
