package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	. "github.com/smartystreets/goconvey/convey"
)

// Logs to an in-memory buffer instead of stdout. This lets the test inspect what was logged
func newTracerWithBuffer() (*dbQueryTracer, *bytes.Buffer) {
	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.DebugLevel)
	return &dbQueryTracer{log: log}, &buf
}

// TestTraceQueryStart_Truncated tests the query tracing so huge statements don't give us grief
func TestTraceQueryStart_Truncated(t *testing.T) {
	Convey("TraceQueryStart truncates long SQL and caps arguments", t, func() {
		tr, _ := newTracerWithBuffer()

		longSQL := strings.Repeat("X", maxSQLLen+123)
		args := make([]any, maxArgsShow+10)
		for i := range args {
			args[i] = i
		}

		ctx := tr.TraceQueryStart(
			context.Background(),
			nil,
			pgx.TraceQueryStartData{
				SQL:  longSQL,
				Args: args,
			},
		)

		v := ctx.Value(traceKey{})
		So(v, ShouldNotBeNil)
		q, ok := v.(*traceQueryData)
		So(ok, ShouldBeTrue)

		// SQL is truncated and ends with ellipsis ("..." is 3 bytes)
		expectedLen := maxSQLLen + len("...")
		So(len(q.sql), ShouldEqual, expectedLen)
		So(strings.HasSuffix(q.sql, "..."), ShouldBeTrue)

		// Args capped and last element is the truncation marker
		So(len(q.args), ShouldEqual, maxArgsShow+1)
		So(fmt.Sprintf("%v", q.args[len(q.args)-1]), ShouldEqual, "...(truncated)")
	})
}

// TestTraceQuery_Warnings tests that we're testing for slow queries
func TestTraceQuery_Warnings(t *testing.T) {
	Convey("TraceQueryEnd logs debug for fast and warn for slow queries", t, func() {
		tr, buf := newTracerWithBuffer()

		ctx := tr.TraceQueryStart(
			context.Background(),
			nil,
			pgx.TraceQueryStartData{
				SQL:  "select 1",
				Args: []any{42},
			},
		)

		// fast (< slowThresh)
		if q, ok := ctx.Value(traceKey{}).(*traceQueryData); ok {
			q.start = time.Now().Add(-slowThresh + 10*time.Millisecond)
		}
		tr.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
			CommandTag: pgconn.CommandTag{},
			Err:        nil,
		})

		out := buf.String()
		So(out, ShouldContainSubstring, `"level":"debug"`)
		So(out, ShouldContainSubstring, "Query finished")

		// slow (>= slowThresh)
		buf.Reset()
		if q, ok := ctx.Value(traceKey{}).(*traceQueryData); ok {
			q.start = time.Now().Add(-slowThresh - 10*time.Millisecond)
		}
		tr.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
			CommandTag: pgconn.CommandTag{},
			Err:        nil,
		})

		out = buf.String()
		So(out, ShouldContainSubstring, `"level":"warn"`)
		So(out, ShouldContainSubstring, "Query finished")
	})
}

// TestTraceQuery_Normalization ensures we handle the empty command when using commandTag
// This ensures observability correctness
func TestTraceQuery_Normalization(t *testing.T) {
	Convey("TraceQueryEnd logs error and normalizes empty CommandTag to QUERY", t, func() {
		tr, buf := newTracerWithBuffer()

		ctx := tr.TraceQueryStart(
			context.Background(),
			nil,
			pgx.TraceQueryStartData{
				SQL:  "delete from x where id=$1",
				Args: []any{99},
			},
		)

		tr.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
			CommandTag: pgconn.CommandTag{},
			Err:        errors.New("boom"),
		})

		out := buf.String()
		So(out, ShouldContainSubstring, `"level":"error"`)
		So(out, ShouldContainSubstring, "Query error")
		So(out, ShouldContainSubstring, `"cmd":"QUERY"`)
		So(out, ShouldContainSubstring, `"sql":"delete from x where id=$1"`)
		So(out, ShouldContainSubstring, `"args":[`)
	})
}
