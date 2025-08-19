package config

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestMustString tests string env vars
func TestMustString(t *testing.T) {
	Convey("MustString returns the value if env var exists", t, func() {
		key := "TEST_KEY_STRING"
		val := "hello"
		os.Setenv(key, val)
		defer os.Unsetenv(key)

		So(MustString(key), ShouldEqual, val)
	})

	Convey("MustString panics if env var is missing", t, func() {
		key := "MISSING_KEY_STRING"
		os.Unsetenv(key)

		So(func() { MustString(key) }, ShouldPanic)
	})
}

// TestMustInt tests int env vars
func TestMustInt(t *testing.T) {
	Convey("MustInt returns an int value if env var is valid", t, func() {
		key := "TEST_KEY_INT"
		os.Setenv(key, "42")
		defer os.Unsetenv(key)

		So(MustInt(key), ShouldEqual, 42)
	})

	Convey("MustInt panics if env var is not set", t, func() {
		key := "MISSING_KEY_INT"
		os.Unsetenv(key)

		So(func() { MustInt(key) }, ShouldPanic)
	})

	Convey("MustInt panics if env var is not an integer", t, func() {
		key := "BAD_KEY_INT"
		os.Setenv(key, "forty-two")
		defer os.Unsetenv(key)

		So(func() { MustInt(key) }, ShouldPanic)
	})
}

// TestMustPort tests ports, and ensures range checking
func TestMustPort(t *testing.T) {
	Convey("MustPort returns :port if env var is valid", t, func() {
		key := "TEST_KEY_PORT"
		os.Setenv(key, "8080")
		defer os.Unsetenv(key)

		So(MustPort(key), ShouldEqual, ":8080")
	})

	Convey("MustPort panics if env var is not set", t, func() {
		key := "MISSING_KEY_PORT"
		os.Unsetenv(key)

		So(func() { MustPort(key) }, ShouldPanic)
	})

	Convey("MustPort panics if env var is not an integer", t, func() {
		key := "BAD_KEY_PORT"
		os.Setenv(key, "abc")
		defer os.Unsetenv(key)

		So(func() { MustPort(key) }, ShouldPanic)
	})

	Convey("MustPort panics if port is out of range", t, func() {
		key := "OUT_OF_RANGE_PORT"
		os.Setenv(key, "70000")
		defer os.Unsetenv(key)

		So(func() { MustPort(key) }, ShouldPanic)
	})
}
