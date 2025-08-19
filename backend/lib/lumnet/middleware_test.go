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

// TestSendRequestID_Missing sets X-Request-Id
func TestSendRequestID_Missing(t *testing.T) {
	Convey("sendRequestID sets X-Request-Id header when RequestID middleware is present", t, func() {
		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		h := middleware.RequestID(sendRequestID(final))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusOK)
		id := rec.Header().Get(requestIDHeader)
		So(id, ShouldNotBeBlank)
	})
}

// TestSendRequestID_Existing doesn't overwrite X-Request-Id
func TestSendRequestID_Existing(t *testing.T) {
	Convey("sendRequestID does not overwrite X-Request-Id if already set", t, func() {
		const preset = "preset-123"

		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		presetHeader := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(requestIDHeader, preset)
			sendRequestID(final).ServeHTTP(w, r)
		})

		h := middleware.RequestID(presetHeader)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusOK)
		So(rec.Header().Get(requestIDHeader), ShouldEqual, preset)
	})
}

// TestRecovery_Panic tests converting a panic to a JSON 500
func TestRecovery_Panic(t *testing.T) {
	Convey("Recovery transforms a panic into a 500 JSON error with ErrorCodePanic", t, func() {
		panicing := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		Recovery(panicing).ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusInternalServerError)

		var body struct {
			StatusCode int    `json:"status_code"`
			Status     string `json:"status"`
			Code       int64  `json:"code"`
			Error      string `json:"error"`
		}
		So(json.Unmarshal(rec.Body.Bytes(), &body), ShouldBeNil)

		So(body.StatusCode, ShouldEqual, http.StatusInternalServerError)
		So(body.Code, ShouldEqual, int64(lumErrors.ErrorCodePanic))
		So(body.Error, ShouldContainSubstring, "internal server error")
	})
}

// TestRecovery_NoPanic tests recovering passes
func TestRecovery_NoPanic(t *testing.T) {
	Convey("Recovery passes through when no panic occurs", t, func() {
		called := false
		ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusTeapot)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		Recovery(ok).ServeHTTP(rec, req)

		So(called, ShouldBeTrue)
		So(rec.Code, ShouldEqual, http.StatusTeapot)
	})
}
