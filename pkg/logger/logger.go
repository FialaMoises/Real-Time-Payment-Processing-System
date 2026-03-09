package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(level string) {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Create logger
	Logger = zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set as global logger
	log.Logger = Logger
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
