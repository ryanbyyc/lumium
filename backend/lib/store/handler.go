package store

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"lumium/lib/config"
	"lumium/lib/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// seams, which are overwritten in tests
var (
	parsePoolConfig   = pgxpool.ParseConfig
	newPoolWithConfig = pgxpool.NewWithConfig
	poolPing          = func(p *pgxpool.Pool, ctx context.Context) error { return p.Ping(ctx) }
	poolClose         = func(p *pgxpool.Pool) { p.Close() }
	timeSleep         = time.Sleep

	// singleton, handler and initialized
	once           sync.Once
	handler        *Handler
	initializedPgx bool
)

// Handler is our database wrapper
type Handler struct {
	PgxPool *pgxpool.Pool
}

// NewHandler creates a new handler and returns it
func NewHandler(initializePgx bool) *Handler {
	once.Do(func() { handler = &Handler{} })
	if initializePgx {
		handler.InitializeDB()
	}
	return handler
}

// GetHandler returns the handler, but panics if the database hasn't spun up
func GetHandler() *Handler {
	if handler == nil {
		panic("store handler not initialized")
	}
	return handler
}

// InitializeDB creates the pgx pool
func (h *Handler) InitializeDB() {
	if initializedPgx && h.PgxPool != nil {
		return
	}

	// Honestly, I pulled these config values out of my hat
	// They could be env variables or part of the config package
	const (
		minConns        = int32(0)
		maxConns        = int32(10)
		maxConnLifetime = 30 * time.Minute
		maxConnIdleTime = 15 * time.Minute
		pingTimeout     = 5 * time.Second
		maxRetries      = 8
		baseBackoff     = 500 * time.Millisecond
		maxBackoff      = 10 * time.Second
	)

	l := logger.Get()

	dsn := config.MustString("SERVICE_PGSQL_DBURL")
	pgcfg, err := parsePoolConfig(dsn)
	if err != nil {
		l.Fatal().Err(err).Msg("pgxpool: could not parse SERVICE_PGSQL_DBURL")
	}

	pgcfg.MinConns = minConns
	pgcfg.MaxConns = maxConns
	pgcfg.MaxConnLifetime = maxConnLifetime
	pgcfg.MaxConnIdleTime = maxConnIdleTime

	// attach tracer so we can view SQL activity
	pgcfg.ConnConfig.Tracer = &dbQueryTracer{log: l}

	pool, err := newPoolWithConfig(context.Background(), pgcfg)
	if err != nil {
		l.Fatal().Err(err).Msg("pgxpool: failed to create pool")
	}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		lastErr = poolPing(pool, ctx)
		cancel()

		if lastErr == nil {
			h.PgxPool = pool
			initializedPgx = true
			l.Info().Int32("max_conns", maxConns).Msg("Database connection established")
			return
		}

		// exponential backoff capped at maxBackoff, with ~20% jitter
		backoff := baseBackoff * (1 << i)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		jitter := time.Duration(rand.Int63n(int64(backoff) / 5))
		sleep := backoff + jitter

		l.Warn().
			Err(lastErr).
			Dur("retry_in", sleep).
			Int("attempt", i+1).
			Int("max", maxRetries+1).
			Msg("Database ping failed; retrying")

		timeSleep(sleep)
	}

	poolClose(pool)
	l.Fatal().Err(lastErr).Msg("pgxpool: could not establish database connectivity")
}

// Close closes the database connection
func (h *Handler) Close() {
	if h.PgxPool != nil {
		poolClose(h.PgxPool) // use seam, not h.PgxPool.Close()
		h.PgxPool = nil
	}
	initializedPgx = false
}
