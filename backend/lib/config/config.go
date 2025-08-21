// Package config is the Lumium config wrapper
package config

import (
	"lumium/lib/logger"
	"os"
	"strconv"
	"strings"
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

// MayString returns the env value or def if unset/empty (trimmed).
func MayString(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

// MayInt returns the int value of the env var, or def if unset/invalid.
func MayInt(key string, def int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		l := logger.Get()
		l.Warn().Str("key", key).Str("value", s).Int("default", def).Msg("Invalid int; using default")
		return def
	}
	return v
}

// MayBool returns the bool value of the env var, or def if unset/invalid.
// Accepts: 1/0, t/true/f/false (case-insensitive)
func MayBool(key string, def bool) bool {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		l := logger.Get()
		l.Warn().Str("key", key).Str("value", s).Bool("default", def).Msg("Invalid bool; using default")
		return def
	}
	return v
}