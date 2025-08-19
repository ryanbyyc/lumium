package config

import (
	"lumium/lib/logger"
	"os"
	"strconv"
)

// MustString expects an environment variable, and panics if it doesn't exist
func MustString(key string) string {
	l := logger.Get()
	v := os.Getenv(key)
	if v == "" {
		l.Panic().Str("key", key).Msg("Unknown key")
	}
	return v
}

// MustInt expects an environment variable, and panics if it doesn't exist or isn't an int
func MustInt(key string) int {
	l := logger.Get()
	s := MustString(key)
	v, err := strconv.Atoi(s)
	if err != nil {
		l.Panic().Str("key", key).Msg("Key value is not an int")
	}
	return v
}

// MustPort expects an environment variable, and panics if it doesn't exist or invalidates
func MustPort(key string) string {
	l := logger.Get()
	s := MustString(key)
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		l.Panic().Str("key", key).Str("value", s).Msg("Invalid TCP port; expected 1..65535")
	}
	return ":" + s
}
