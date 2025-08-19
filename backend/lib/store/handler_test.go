package store

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/smartystreets/goconvey/convey"
)

// reset global singletons & seams between tests
func resetHandlerSeams() {
	// reset singletons/state
	once = sync.Once{}
	handler = nil
	initializedPgx = false

	// reset seams to real values
	parsePoolConfig = pgxpool.ParseConfig
	newPoolWithConfig = pgxpool.NewWithConfig
	poolPing = func(p *pgxpool.Pool, ctx context.Context) error { return p.Ping(ctx) }
	poolClose = func(p *pgxpool.Pool) { p.Close() }
	timeSleep = time.Sleep
}

// TestNewHandler tests a new singleton
func TestNewHandler(t *testing.T) {
	Convey("NewHandler(false) returns singleton and does not initialize pgx", t, func() {
		resetHandlerSeams()
		t.Setenv("DATABASE_URL", "postgres://ignore@localhost/db")

		h1 := NewHandler(false)
		So(h1, ShouldNotBeNil)
		So(h1.PgxPool, ShouldBeNil)
		So(initializedPgx, ShouldBeFalse)

		h2 := NewHandler(false)
		So(h2, ShouldEqual, h1) // singleton
	})
}

// TestInitializeDB tries to initialize the database and setup
func TestInitializeDB(t *testing.T) {
	Convey("InitializeDB succeeds and sets pool + initializedPgx", t, func() {
		resetHandlerSeams()
		t.Setenv("DATABASE_URL", "postgres://ignore@localhost/db")

		// stub: parse config returns a minimal usable config
		parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
			cfg := &pgxpool.Config{}
			cfg.ConnConfig = &pgx.ConnConfig{} // minimal
			return cfg, nil
		}

		// stub: new pool just returns a zero pool pointer
		// we will never call real methods in testing
		newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
			return &pgxpool.Pool{}, nil
		}

		// stub: ping returns nil immediately
		poolPing = func(_ *pgxpool.Pool, _ context.Context) error { return nil }

		// no sleeping
		timeSleep = func(time.Duration) {}

		h := NewHandler(true)
		So(h, ShouldNotBeNil)
		So(h.PgxPool, ShouldNotBeNil)
		So(initializedPgx, ShouldBeTrue)

		// GetHandler should return the same singleton
		So(GetHandler(), ShouldEqual, h)
	})
}

// TestInitializeDB_Retry tests our retry logic
func TestInitializeDB_Retry(t *testing.T) {
	Convey("InitializeDB retries on ping failures and succeeds", t, func() {
		resetHandlerSeams()
		t.Setenv("DATABASE_URL", "postgres://ignore@localhost/db")

		parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
			cfg := &pgxpool.Config{ConnConfig: &pgx.ConnConfig{}}
			return cfg, nil
		}
		newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
			return &pgxpool.Pool{}, nil
		}

		call := 0
		poolPing = func(_ *pgxpool.Pool, _ context.Context) error {
			call++
			if call < 3 {
				return context.DeadlineExceeded
			}
			return nil
		}
		timeSleep = func(time.Duration) {} // skip waiting

		h := NewHandler(true)
		So(h, ShouldNotBeNil)
		So(initializedPgx, ShouldBeTrue)
		So(h.PgxPool, ShouldNotBeNil)
		So(call, ShouldBeGreaterThanOrEqualTo, 3) // retried at least twice before success
	})
}

// TestClose test close on the DB
func TestClose(t *testing.T) {
	Convey("Close() closes the pool and resets state", t, func() {
		resetHandlerSeams()
		t.Setenv("DATABASE_URL", "postgres://ignore@localhost/db")

		parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
			return &pgxpool.Config{ConnConfig: &pgx.ConnConfig{}}, nil
		}
		newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
			return &pgxpool.Pool{}, nil
		}
		poolPing = func(_ *pgxpool.Pool, _ context.Context) error { return nil }
		timeSleep = func(time.Duration) {}

		var closed bool
		poolClose = func(_ *pgxpool.Pool) { closed = true } // stub closes safely

		h := NewHandler(true)
		So(initializedPgx, ShouldBeTrue)
		So(h.PgxPool, ShouldNotBeNil)

		h.Close()
		So(closed, ShouldBeTrue)
		So(h.PgxPool, ShouldBeNil)
		So(initializedPgx, ShouldBeFalse)
	})
}

// TestInitializeDB_Fatal handles database url absence which would normally fatal out
// We don't trigger Fatal in tests
func TestInitializeDB_Fatal(t *testing.T) {
	Convey("Guard: with seams in place, we never hit l.Fatal() in tests", t, func() {
		resetHandlerSeams()
		// Provide a URL so ParseConfig seam is invoked (we don't want to test Fatal path)
		t.Setenv("DATABASE_URL", "postgres://ignore@localhost/db")

		parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
			return &pgxpool.Config{ConnConfig: &pgx.ConnConfig{}}, nil
		}
		newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
			return &pgxpool.Pool{}, nil
		}
		poolPing = func(_ *pgxpool.Pool, _ context.Context) error { return nil }

		h := NewHandler(true)
		So(h, ShouldNotBeNil)
	})
}
