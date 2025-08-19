package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type (
	dbQueryTracer struct {
		log zerolog.Logger
	}
	traceQueryData struct {
		start time.Time
		sql   string
		args  []any
	}
	traceKey struct{}
)

const (
	maxSQLLen   = 4000                   // truncate long statements
	maxArgsShow = 50                     // cap args logged
	slowThresh  = 500 * time.Millisecond // half a second query time, seems reasonable
)

// TraceQueryStart adds debugging information to the context
func (t *dbQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	sql := data.SQL
	if len(sql) > maxSQLLen {
		sql = sql[:maxSQLLen] + "..."
	}

	args := data.Args
	if len(args) > maxArgsShow {
		args = append(args[:maxArgsShow], "...(truncated)")
	}

	q := &traceQueryData{
		start: time.Now(),
		sql:   sql,
		args:  args,
	}

	return context.WithValue(ctx, traceKey{}, q)
}

// TraceQueryEnd takes what we had in TraceQueryStart and adds it to the log.
func (t *dbQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	v := ctx.Value(traceKey{})
	q, _ := v.(*traceQueryData) // may be nil

	// Safe defaults
	var (
		elapsed time.Duration
		sql     string
		args    any
	)

	if q != nil {
		elapsed = time.Since(q.start)
		sql = compactSQL(q.sql)
		args = q.args
	}

	ev := t.log.With().Dur("elapsed", elapsed).Logger()

	// Normalize command label
	cmd := strings.ToUpper(strings.TrimSpace(data.CommandTag.String()))
	if cmd == "" {
		cmd = "QUERY"
	}

	evt := ev.Debug()
	// warn for slow queries (only meaningful if we captured a start time)
	if slowThresh > 0 && elapsed >= slowThresh {
		evt = ev.Warn()
	}

	if data.Err != nil {
		ev.Error().
			Err(data.Err).
			Str("cmd", cmd).
			Str("sql", sql).
			Any("args", args).
			Msg("Query error")
		return
	}

	evt.
		Str("cmd", cmd).
		Str("sql", sql).
		Any("args", args).
		Msg("Query finished")
}

// compactSQL flattens all whitespace (including \n and \t) to single spaces
func compactSQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
