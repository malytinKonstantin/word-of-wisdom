package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	zerolog.TimeFieldFormat = time.RFC3339
	Log = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}
