package logger

import (
	"os"
	"testing"
	"github.com/rs/zerolog"
)

func TestInit(t *testing.T) {
	// Test development mode
	os.Setenv("ENV", "development")
	os.Setenv("LOG_LEVEL", "info")
	
	Init()
	
	if Log.GetLevel() != zerolog.InfoLevel {
		t.Error("Log level should be info")
	}
}

func TestInitDebugLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	Init()
	
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Error("Global level should be debug")
	}
}

func TestInitWarnLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "warn")
	Init()
	
	if zerolog.GlobalLevel() != zerolog.WarnLevel {
		t.Error("Global level should be warn")
	}
}

func TestInitErrorLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "error")
	Init()
	
	if zerolog.GlobalLevel() != zerolog.ErrorLevel {
		t.Error("Global level should be error")
	}
}

func TestInitDefaultLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "")
	Init()
	
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Error("Default level should be info")
	}
}

func TestInitProductionMode(t *testing.T) {
	os.Setenv("ENV", "production")
	Init()
	
	// Should not panic
}

func TestHelperFunctions(t *testing.T) {
	Init()
	
	// Test that helper functions don't panic
	Info().Msg("test")
	Debug().Msg("test")
	Warn().Msg("test")
	Error().Msg("test")
	
	// Note: Fatal() would exit the program, so we don't test it
}
