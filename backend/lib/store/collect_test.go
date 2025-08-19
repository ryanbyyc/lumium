package store

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Fake pgx.Rows for unit tests
type (
	fakeRows struct {
		cols      []string
		data      [][]any
		i         int
		closed    bool
		scanErrAt int // if > 0, return error on that row
	}
	recSimple struct {
		ID   int64
		Name string
	}
)

// newFakeRows makes faux data of *fakeRows
func newFakeRows(cols []string, data [][]any) *fakeRows {
	return &fakeRows{cols: cols, data: data, i: -1}
}

// Close closes the fakeRows
func (r *fakeRows) Close() { r.closed = true }

// Err is the fakeRows Error
func (r *fakeRows) Err() error { return nil }

// CommandTag is the seam for fakeRows
func (r *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// Conn is the seam for fakeRows
func (r *fakeRows) Conn() *pgx.Conn { return nil }

// RawValues is the seam for fakeRows
func (r *fakeRows) RawValues() [][]byte { return nil }

// Next is the seam for fakeRows
func (r *fakeRows) Next() bool { r.i++; return r.i < len(r.data) }

// Values is the seam for fakeRows
func (r *fakeRows) Values() ([]any, error) {
	if r.i < 0 || r.i >= len(r.data) {
		return nil, errors.New("out of range")
	}
	return r.data[r.i], nil
}

// FieldDescriptions is the seam for fakeRows
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	fds := make([]pgconn.FieldDescription, len(r.cols))
	for i, c := range r.cols {
		fds[i] = pgconn.FieldDescription{Name: c}
	}
	return fds
}

// Scan is the seam for fakeRows
func (r *fakeRows) Scan(dest ...any) error {
	// simulate error on specific row, if requested
	if r.scanErrAt > 0 && (r.i+1) == r.scanErrAt {
		return errors.New("scan error")
	}
	row := r.data[r.i]
	if len(dest) != len(row) {
		return errors.New("dest length mismatch")
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *int64:
			v, _ := row[i].(int64)
			*d = v
		case *int32:
			v, _ := row[i].(int32)
			*d = v
		case *float32:
			v, _ := row[i].(float32)
			*d = v
		case *string:
			v, _ := row[i].(string)
			*d = v
		default:
			return errors.New("unsupported scan dest type")
		}
	}
	return nil
}

// TestCollectStructsByName_Success tests a successful collection
func TestCollectStructsByName_Success(t *testing.T) {
	Convey("CollectStructsByName maps columns to struct fields by name", t, func() {
		rows := newFakeRows(
			[]string{"ID", "Name"}, // must match field names
			[][]any{
				{int64(1), "alpha"},
				{int64(2), "beta"},
			},
		)

		got, err := CollectStructsByName[recSimple](rows)
		So(err, ShouldBeNil)
		So(got, ShouldResemble, []recSimple{
			{ID: 1, Name: "alpha"},
			{ID: 2, Name: "beta"},
		})
	})
}

// TestCollectStructsByName_Error triggers an error
func TestCollectStructsByName_Error(t *testing.T) {
	Convey("CollectStructsByName wraps scan errors as DB errors", t, func() {
		r := newFakeRows(
			[]string{"ID", "Name"},
			[][]any{
				{int64(1), "ok"},
				{int64(2), "boom"},
			},
		)
		r.scanErrAt = 2 // force error on row 2

		_, err := CollectStructsByName[recSimple](r)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to collect rows")
	})
}
