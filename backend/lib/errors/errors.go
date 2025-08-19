package errors

import (
	sterrors "errors" // alias stdlib
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// Error represents an error that could be wrapping another error, it includes a code for determining what
// triggered the error
type Error struct {
	orig  error
	msg   string
	code  ErrorCode
	field string
}

type Wire struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Field   string    `json:"field,omitempty"`
}

// ErrorCode defines supported error codes (iota)
type ErrorCode uint

const (
	// ErrorCodeUnknown is the default error code
	ErrorCodeUnknown ErrorCode = iota

	// ErrorCodeNotFound is the 404 error code
	ErrorCodeNotFound

	// ErrorCodeInvalidArgument is the bad request
	ErrorCodeInvalidArgument

	// ErrorCodeDuplicateKey is the duplicate key error
	ErrorCodeDuplicateKey

	// ErrorCodeDB is a general error code for database errors
	ErrorCodeDB

	// ErrorCodeValidation is a general error code for validation errors
	ErrorCodeValidation

	// ErrorCodeJSON is the general error code for JSON related errors
	ErrorCodeJSON

	// ErrorCodePanic is the general error code for panics
	ErrorCodePanic
)

// example of how I would handle specific error codes
// such as foreign_key_violation
const (
	errDuplicateKey = "23505"
)

// HTTPStatusCode turns an ErrorCode to an http status code
func HTTPStatusCode(s ErrorCode) int {
	switch s {
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeInvalidArgument:
		return http.StatusUnprocessableEntity
	case ErrorCodeDuplicateKey:
		return http.StatusConflict
	case ErrorCodeValidation:
		return http.StatusBadRequest
	case ErrorCodeDB, ErrorCodeJSON, ErrorCodePanic, ErrorCodeUnknown:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NotFoundf is a convenience method for not found errors
func NotFoundf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodeNotFound, format, a...)
}

// InvalidArgf is a convenience method for invalid arguments
func InvalidArgf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodeInvalidArgument, format, a...)
}

// DuplicateKeyf is a convenience method for duplicate keys
func DuplicateKeyf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodeDuplicateKey, format, a...)
}

// DBf is a convenience method for a general database error
func DBf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodeDB, format, a...)
}

// JSONErrf is a convenience method for a general json error
func JSONErrf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodeJSON, format, a...)
}

// PanicErrf is a convenience method for a panic
func PanicErrf(format string, a ...interface{}) error {
	return NewErrorf(ErrorCodePanic, format, a...)
}

// WithField chains an error and sets the field
func (e *Error) WithField(field string) *Error { e.field = field; return e }

// WithCode chains an error and sets the ErrorCode
func (e *Error) WithCode(code ErrorCode) *Error { e.code = code; return e }

// WithCause chains an error and sets the cause (original error)
func (e *Error) WithCause(orig error) *Error { e.orig = orig; return e }

// ToWire returns a new wired error code
func (e *Error) ToWire() Wire {
	return Wire{Code: e.code, Message: e.msg, Field: e.field}
}

// WrapErrorf returns a wrapped error
func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		code: code,
		orig: orig,
		msg:  fmt.Sprintf(format, a...),
	}
}

// NewErrorf instantiates a new error
func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}

// NewValidationError instantiates a new error with details of the offending field
func NewValidationError(code ErrorCode, msg string, field string) error {
	return &Error{
		code:  code,
		orig:  nil,
		msg:   msg,
		field: field,
	}
}

// Error returns the message, when wrapping errors the wrapped error is returned.
func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// IsErrorCode is a helper for allowing us to compare error states
func IsErrorCode(err error, code ErrorCode) bool {
	var e *Error
	if sterrors.As(err, &e) {
		return e.code == code
	}
	return false
}

// Unwrap returns the wrapped error, if any
func (e *Error) Unwrap() error {
	return e.orig
}

// Code returns the code representing this error
func (e *Error) Code() ErrorCode {
	return e.code
}

// Field returns the code representing this error
func (e *Error) Field() string {
	return e.field
}

// DBErrorCode accepts an int code and then tries to convert it to an error code
// Returns nil if not found to allow for edge case handling
func DBErrorCode(err error) *ErrorCode {
	var pgErr *pgconn.PgError
	if sterrors.As(err, &pgErr) {
		switch pgErr.Code {
		case errDuplicateKey:
			c := ErrorCodeDuplicateKey
			return &c
		}
	}
	return nil
}
