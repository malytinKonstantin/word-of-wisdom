package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	level := zerolog.InfoLevel
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	}
	multiWriter := io.MultiWriter(consoleWriter)
	Log = zerolog.New(multiWriter).Level(level).With().Timestamp().Logger()
}
