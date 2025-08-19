package lumnet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lumErrors "lumium/lib/errors"

	"github.com/go-chi/chi/v5/middleware"
	. "github.com/smartystreets/goconvey/convey"
)

// TestRenderError responds 500 & ensures request ID is set
func TestRenderError(t *testing.T) {
	Convey("RenderError with defacto error returns 500 and includes request id", t, func() {
		// a handler that just calls RenderError
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			RenderError(w, r, assertErr("boom"))
		})

		// add RequestID so request_id field is set
		wrapped := middleware.RequestID(h)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		wrapped.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusInternalServerError)

		var body struct {
			StatusCode int    `json:"status_code"`
			Status     string `json:"status"`
			Code       int64  `json:"code"`
			Error      string `json:"error"`
			RequestID  string `json:"request_id"`
			Payload    any    `json:"payload"`
		}
		So(json.Unmarshal(rec.Body.Bytes(), &body), ShouldBeNil)
		So(body.StatusCode, ShouldEqual, http.StatusInternalServerError)
		So(body.Status, ShouldEqual, http.StatusText(http.StatusInternalServerError))
		So(body.Code, ShouldEqual, int64(lumErrors.ErrorCodeUnknown))
		So(body.Error, ShouldContainSubstring, "boom")
		So(body.RequestID, ShouldNotBeBlank)
		So(body.Payload, ShouldBeNil)
	})
}

// TestRenderError_CommonValidation maps a validation to 400
func TestRenderError_CommonValidation(t *testing.T) {
	Convey("RenderError maps commonErrors validation to 400 with field and payload", t, func() {
		err := lumErrors.NewValidationError(lumErrors.ErrorCodeValidation, "is required", "name")

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			RenderError(w, r, err)
		})
		wrapped := middleware.RequestID(h)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		wrapped.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusBadRequest)

		var body struct {
			StatusCode         int    `json:"status_code"`
			Status             string `json:"status"`
			Code               int64  `json:"code"`
			Error              string `json:"error"`
			ValidationErrorKey string `json:"validation_field"`
			RequestID          string `json:"request_id"`
			Payload            struct {
				Code    int64  `json:"code"`
				Message string `json:"message"`
				Field   string `json:"field"`
			} `json:"payload"`
		}
		So(json.Unmarshal(rec.Body.Bytes(), &body), ShouldBeNil)

		So(body.StatusCode, ShouldEqual, http.StatusBadRequest)
		So(body.Status, ShouldEqual, http.StatusText(http.StatusBadRequest))
		So(body.Code, ShouldEqual, int64(lumErrors.ErrorCodeValidation))
		So(body.Error, ShouldContainSubstring, "is required")
		So(body.ValidationErrorKey, ShouldEqual, "name")
		So(body.RequestID, ShouldNotBeBlank)

		// emsire payload mirrors the error wire shape
		So(body.Payload.Code, ShouldEqual, int64(lumErrors.ErrorCodeValidation))
		So(body.Payload.Message, ShouldContainSubstring, "is required")
		So(body.Payload.Field, ShouldEqual, "name")
	})
}

// TestOK tests 200 w/ JSON body
func TestOK(t *testing.T) {
	Convey("OK returns 200 and JSON body", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		payload := map[string]any{"k": "v"}
		OK(rec, req, payload)

		So(rec.Code, ShouldEqual, http.StatusOK)

		var got map[string]any
		So(json.Unmarshal(rec.Body.Bytes(), &got), ShouldBeNil)
		So(got["k"], ShouldEqual, "v")
	})
}

// TestCreated returns a 201 and a URL
// Alternatively, we could have responded with a payload
func TestCreated(t *testing.T) {
	Convey("Created sets 201 and Location header", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		payload := map[string]any{"id": "123"}
		Created(rec, req, payload, "/books/123")

		So(rec.Code, ShouldEqual, http.StatusCreated)
		So(rec.Header().Get("Location"), ShouldEqual, "/books/123")

		var got map[string]any
		So(json.Unmarshal(rec.Body.Bytes(), &got), ShouldBeNil)
		So(got["id"], ShouldEqual, "123")
	})
}

// TestNoContent responds 204 (no content)
func TestNoContent(t *testing.T) {
	Convey("NoContent returns 204 and no body", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)

		NoContent(rec, req)

		So(rec.Code, ShouldEqual, http.StatusNoContent)
		So(rec.Body.Len(), ShouldEqual, 0)
	})
}

// TestData wraps a payload `{data: any}“
func TestData(t *testing.T) {
	Convey("Data wraps payload under data", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		Data(rec, req, map[string]any{"x": 1})

		So(rec.Code, ShouldEqual, http.StatusOK)
		var got struct {
			Data map[string]any `json:"data"`
		}
		So(json.Unmarshal(rec.Body.Bytes(), &got), ShouldBeNil)
		So(got.Data["x"], ShouldEqual, float64(1)) // JSON numbers are float64
	})
}

// TestList is a test for a list sample payload
// This really could look like anything.
func TestList(t *testing.T) {
	Convey("List returns items/total/meta", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		items := []int{1, 2, 3}
		meta := map[string]any{"page": 1}
		List(rec, req, items, 42, meta)

		So(rec.Code, ShouldEqual, http.StatusOK)
		var got struct {
			Items []int          `json:"items"`
			Total int64          `json:"total"`
			Meta  map[string]any `json:"meta"`
		}
		So(json.Unmarshal(rec.Body.Bytes(), &got), ShouldBeNil)
		So(got.Items, ShouldResemble, []int{1, 2, 3})
		So(got.Total, ShouldEqual, 42)
		So(got.Meta["page"], ShouldEqual, float64(1))
	})
}

// TestJSONAny tests convenience methods
func TestJSONAny(t *testing.T) {
	Convey("JSONAny returns error via RenderError", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		JSONAny(rec, req, nil, lumErrors.NotFoundf("missing"))
		So(rec.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("JSONAny POST non-nil should be 201", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		JSONAny(rec, req, map[string]any{"z": 9}, nil)
		So(rec.Code, ShouldEqual, http.StatusCreated)
	})

	Convey("JSONAny GET nil should be 204", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		JSONAny(rec, req, nil, nil)
		So(rec.Code, ShouldEqual, http.StatusNoContent)
	})

	Convey("JSONAny PUT nil should be 200", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		JSONAny(rec, req, nil, nil)
		So(rec.Code, ShouldEqual, http.StatusOK)
	})
}

// TestJSONStatus tests when we set status
func TestJSONStatus(t *testing.T) {
	Convey("JSONStatus writes deliberate status", t, func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		JSONStatus(rec, req, map[string]any{"ok": true}, http.StatusAccepted)

		So(rec.Code, ShouldEqual, http.StatusAccepted)
		var got map[string]any
		So(json.Unmarshal(rec.Body.Bytes(), &got), ShouldBeNil)
		So(got["ok"], ShouldEqual, true)
	})
}

// helpers
func assertErr(msg string) error { return &simpleErr{msg: msg} }

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }
