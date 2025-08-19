package logger

import (
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var once sync.Once
var log zerolog.Logger

// Get returns the logger instance as a singleton
func Get() zerolog.Logger {
	once.Do(func() {

		// Zerolog global configuration
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		// Determine log level from env (names or numbers), default to DEBUG
		level := levelFromEnvDefaultDebug()

		// Console writer to stdout (human-friendly for dev)
		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		// Safe build info
		buildInfo, ok := debug.ReadBuildInfo()
		goVersion := "unknown"
		if ok && buildInfo != nil {
			goVersion = buildInfo.GoVersion
		}

		log = zerolog.New(output).
			Level(level).
			With().
			Timestamp().
			Str("go_version", goVersion).
			Logger()
	})

	return log
}

// levelFromEnvDefaultDebug parses LOG_LEVEL from the environment
// Supports named levels ("debug", "info", "warn", "error", "trace", etc.) and numeric levels
// Falls back to DEBUG if missing/invalid/out-of-range
func levelFromEnvDefaultDebug() zerolog.Level {
	raw := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if raw == "" {
		return zerolog.DebugLevel
	}

	// First try named levels
	if lvl, err := zerolog.ParseLevel(strings.ToLower(raw)); err == nil {
		return lvl
	}

	// Then try numeric levels
	if n, err := strconv.Atoi(raw); err == nil {
		lvl := zerolog.Level(n)
		// Basic sanity clamp: known zerolog levels range from TraceLevel..PanicLevel and special NoLevel/Disabled
		if int(lvl) >= int(zerolog.TraceLevel) && int(lvl) <= int(zerolog.PanicLevel) {
			return lvl
		}
	}

	// Fallback
	return zerolog.DebugLevel
}
