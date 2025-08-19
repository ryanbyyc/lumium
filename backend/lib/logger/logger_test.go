package logger

import (
	"sync"
	"testing"

	"github.com/rs/zerolog"
	. "github.com/smartystreets/goconvey/convey"
)

// reset the package-level singleton between tests
func resetLoggerForTest() {
	once = sync.Once{}
	log = zerolog.Logger{}
}

// TestLogger_Defaults tests LOG_LEVEL env variable
func TestLogger_Defaults(t *testing.T) {
	Convey("Get() defaults to DEBUG when LOG_LEVEL is missing", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "") // set to missing

		l := Get()
		So(l.GetLevel(), ShouldEqual, zerolog.DebugLevel)
	})

	Convey("Get() defaults to DEBUG when LOG_LEVEL is not a valid name/number", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "nesto")

		l := Get()
		So(l.GetLevel(), ShouldEqual, zerolog.DebugLevel)
	})
}

// TestLogger_Levels tests levels setup
func TestLogger_Levels(t *testing.T) {
	Convey("LOG_LEVEL=info (by name) sets INFO level", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "info")

		l := Get()
		So(l.GetLevel(), ShouldEqual, zerolog.InfoLevel)
	})

	Convey("LOG_LEVEL=2 (by number) sets WARN level", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "2")

		l := Get()
		So(l.GetLevel(), ShouldEqual, zerolog.WarnLevel)
	})

	Convey("Out-of-range numeric level is honored by zerolog", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "99")

		l := Get()
		// zerolog allows arbitrary levels
		So(l.GetLevel(), ShouldEqual, zerolog.Level(99))
	})
}

// TestLogger_Singleton tests sync.once
func TestLogger_Singleton(t *testing.T) {
	Convey("Get() initializes once and subsequent calls return same configured logger", t, func() {
		resetLoggerForTest()
		t.Setenv("LOG_LEVEL", "warn")

		l1 := Get()
		// Change env after first call; should have no effect
		t.Setenv("LOG_LEVEL", "debug")
		l2 := Get()

		So(l1.GetLevel(), ShouldEqual, zerolog.WarnLevel)
		So(l2.GetLevel(), ShouldEqual, zerolog.WarnLevel)
		So(l1, ShouldResemble, l2)
	})
}
