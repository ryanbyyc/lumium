package lumnet

import (
	commonErrors "lumium/lib/errors"

	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// These conveniece methods assume we're dealing with JSON, and we want to keep our handlers as
// lean as possible. They could be improved by extending them for different response types

// RenderError converts any error to a normalized JSON error and writes the proper status
func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	resp := ErrResponse{
		HTTPStatusCode: status,
		StatusText:     http.StatusText(status),
		AppCode:        int64(commonErrors.ErrorCodeUnknown),
		ErrorText:      err.Error(),
		RequestID:      middleware.GetReqID(r.Context()),
	}

	var ierr *commonErrors.Error
	if As(err, &ierr) {
		status = commonErrors.HTTPStatusCode(ierr.Code())
		resp.HTTPStatusCode = status
		resp.StatusText = http.StatusText(status)
		resp.AppCode = int64(ierr.Code())
		resp.ErrorText = ierr.Error()
		if ierr.Code() == commonErrors.ErrorCodeValidation {
			resp.ValidationErrorField = ierr.Field()
		}
		wire := ierr.ToWire()
		resp.Payload = &wire
	}
	render.Status(r, status)
	render.JSON(w, r, &resp)
}

// OK writes 200 with any JSON marshaled payload
func OK(w http.ResponseWriter, r *http.Request, v any) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, v)
}

// Created writes 201 and optionally sets a Location header
// As mentioned in the tests, we could have alternatively set a response payload from the DB
func Created(w http.ResponseWriter, r *http.Request, v any, location string) {
	if location != "" {
		w.Header().Set("Location", location)
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, v)
}

// NoContent writes 204 with no body
func NoContent(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// Data wraps a payload under {"data": ... }
func Data(w http.ResponseWriter, r *http.Request, v any) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{"data": v})
}

// List writes a common list envelope - we'd use this for things like pagination
func List[T any](w http.ResponseWriter, r *http.Request, items []T, total int64, meta map[string]any) {
	if meta == nil {
		meta = map[string]any{}
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{
		"items": items,
		"total": total,
		"meta":  meta,
	})
}

// JSONAny returns error via RenderError, or a sensible status for payloads
// POST/PUT/PATCH to 201 if payload non-nil, nil to 204, else 200
func JSONAny(w http.ResponseWriter, r *http.Request, p any, err error) {
	if err != nil {
		RenderError(w, r, err)
		return
	}
	status := http.StatusOK
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if p != nil {
			status = http.StatusCreated
		}
	default:
		if p == nil {
			status = http.StatusNoContent
		}
	}
	if p == nil {
		w.WriteHeader(status)
		return
	}
	render.Status(r, status)
	render.JSON(w, r, p)
}

// JSONStatus writes a specific status and payload
func JSONStatus(w http.ResponseWriter, r *http.Request, p map[string]interface{}, status int) {
	render.Status(r, status)
	render.JSON(w, r, p)
}
