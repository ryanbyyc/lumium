package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestHTTPStatusCode tests the expected HTTP status code with API specific error codes
func TestHTTPStatusCode(t *testing.T) {
	cases := []struct {
		code ErrorCode
		want int
	}{
		{ErrorCodeUnknown, http.StatusInternalServerError},
		{ErrorCodeNotFound, http.StatusNotFound},
		{ErrorCodeInvalidArgument, http.StatusUnprocessableEntity},
		{ErrorCodeDuplicateKey, http.StatusConflict},
		{ErrorCodeDB, http.StatusInternalServerError},
		{ErrorCodeValidation, http.StatusBadRequest},
		{ErrorCodeJSON, http.StatusInternalServerError},
		{ErrorCodePanic, http.StatusInternalServerError},
	}
	for _, c := range cases {
		if got := HTTPStatusCode(c.code); got != c.want {
			t.Fatalf("HTTPStatusCode(%v) = %d, want %d", c.code, got, c.want)
		}
	}
}

// TestWrapNewAndErrorFormatting tests error setup and convenience functions
func TestWrapNewAndErrorFormatting(t *testing.T) {
	orig := fmt.Errorf("db went kablooey")
	err := WrapErrorf(orig, ErrorCodeDB, "while saving id=%d", 42)

	// Error string should include both message and wrapped error
	wantSub1 := "while saving id=42"
	wantSub2 := "db went kablooey"
	got := err.Error()
	if !contains(got, wantSub1) || !contains(got, wantSub2) {
		t.Fatalf("Error() = %q, want to contain %q and %q", got, wantSub1, wantSub2)
	}

	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *errors.Error via errors.As")
	}
	if e.Code() != ErrorCodeDB {
		t.Fatalf("Code() = %v, want %v", e.Code(), ErrorCodeDB)
	}
	if e.Unwrap() == nil || e.Unwrap().Error() != "db went kablooey" {
		t.Fatalf("Unwrap() mismatch, got %#v", e.Unwrap())
	}

	// ToWire code/message/field
	w := e.ToWire()
	if w.Code != ErrorCodeDB || !contains(w.Message, "while saving") || w.Field != "" {
		t.Fatalf("ToWire() mismatch: %+v", w)
	}
}

// TestNewValidationError_Field tests validations
func TestNewValidationError_Field(t *testing.T) {
	err := NewValidationError(ErrorCodeValidation, "must be positive", "amount")

	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *errors.Error")
	}
	if e.Code() != ErrorCodeValidation {
		t.Fatalf("code = %v, want %v", e.Code(), ErrorCodeValidation)
	}
	if e.Field() != "amount" {
		t.Fatalf("field = %q, want %q", e.Field(), "amount")
	}
	if e.Error() != "must be positive" {
		t.Fatalf("Error() = %q, want %q", e.Error(), "must be positive")
	}
}

// TestIsErrorCode tests assertions for error codes
func TestIsErrorCode(t *testing.T) {
	base := NewErrorf(ErrorCodeNotFound, "widget %d", 7)
	if !IsErrorCode(base, ErrorCodeNotFound) {
		t.Fatalf("IsErrorCode should be true for matching code")
	}
	if IsErrorCode(base, ErrorCodeDuplicateKey) {
		t.Fatalf("IsErrorCode should be false for non-matching code")
	}
	if IsErrorCode(fmt.Errorf("plain error"), ErrorCodeUnknown) {
		t.Fatalf("IsErrorCode should be false for non-*errors.Error")
	}
}

// TestWithHelpers tests convenience methods
func TestWithHelpers(t *testing.T) {
	// Builders
	e := (&Error{}).WithCode(ErrorCodeJSON).WithField("payload").WithCause(fmt.Errorf("bad json"))
	if e.Code() != ErrorCodeJSON || e.Field() != "payload" || e.Unwrap() == nil {
		t.Fatalf("With* chain mismatch: %+v", e)
	}

	// Constructors
	if !IsErrorCode(NotFoundf("nope"), ErrorCodeNotFound) {
		t.Fatalf("NotFoundf should set ErrorCodeNotFound")
	}
	if !IsErrorCode(InvalidArgf("bad"), ErrorCodeInvalidArgument) {
		t.Fatalf("InvalidArgf should set ErrorCodeInvalidArgument")
	}
	if !IsErrorCode(DuplicateKeyf("dupe"), ErrorCodeDuplicateKey) {
		t.Fatalf("DuplicateKeyf should set ErrorCodeDuplicateKey")
	}
	if !IsErrorCode(DBf("db x"), ErrorCodeDB) {
		t.Fatalf("DBf should set ErrorCodeDB")
	}
	if !IsErrorCode(JSONErrf("json x"), ErrorCodeJSON) {
		t.Fatalf("JSONErrf should set ErrorCodeJSON")
	}
	if !IsErrorCode(PanicErrf("panic x"), ErrorCodePanic) {
		t.Fatalf("PanicErrf should set ErrorCodePanic")
	}
}

// TestDBErrorCode tests currently implemented foreign_key_violation
func TestDBErrorCode(t *testing.T) {
	dupe := &pgconn.PgError{Code: errDuplicateKey}
	other := &pgconn.PgError{Code: "23503"}

	// Wrapped duplicate key
	wrappedDupe := fmt.Errorf("wrap: %w", dupe)
	if c := DBErrorCode(wrappedDupe); c == nil || *c != ErrorCodeDuplicateKey {
		t.Fatalf("DBErrorCode(duplicate) = %v, want %v", deref(c), ErrorCodeDuplicateKey)
	}

	// Wrapped non-mapped error to nil
	wrappedOther := fmt.Errorf("wrap: %w", other)
	if c := DBErrorCode(wrappedOther); c != nil {
		t.Fatalf("DBErrorCode(other) = %v, want nil", *c)
	}

	// Non-pg error to nil
	if c := DBErrorCode(fmt.Errorf("no pg error")); c != nil {
		t.Fatalf("DBErrorCode(non-pg) = %v, want nil", *c)
	}
}

// helper for contains
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}

// helper for indexOf
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// helper for rereferencing
func deref(c *ErrorCode) ErrorCode {
	if c == nil {
		return ErrorCode(999)
	}
	return *c
}
