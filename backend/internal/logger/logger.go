package logger

import (
	"os"
	"time"
	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	// Pretty logging for development
	if os.Getenv("ENV") == "development" {
		Log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		// JSON logging for production
		Log = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	// Set global log level
	level := os.Getenv("LOG_LEVEL")
	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	// Also pin the level on the logger instance so Log.GetLevel() reflects the
	// configured level (not just the process-global level).
	Log = Log.Level(logLevel)

	Log.Info().Msg("Logger initialized")
}

// Helper functions
func Info() *zerolog.Event {
	return Log.Info()
}

func Debug() *zerolog.Event {
	return Log.Debug()
}

func Warn() *zerolog.Event {
	return Log.Warn()
}

func Error() *zerolog.Event {
	return Log.Error()
}

func Fatal() *zerolog.Event {
	return Log.Fatal()
}
