package lumnet

import (
	"bytes"
	"io"
	lumErrors "lumium/lib/errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	. "github.com/smartystreets/goconvey/convey"
)

// testIn is our sample struct for validating
type testIn struct {
	Name string `json:"name" validate:"required,min=2"`
	Age  int    `json:"age"  validate:"gte=1,lte=10"`
}

// TestInit tests the InitValidator singleton
func TestInit(t *testing.T) {
	Convey("InitValidator is a singleton and GetValidator returns the same instance", t, func() {
		v1 := InitValidator()
		v2 := GetValidator()
		So(v1, ShouldNotBeNil)
		So(v2, ShouldNotBeNil)
		So(v1, ShouldEqual, v2)
	})
}

// TestParseJSON_Success tests a successful unmarshal
func TestParseJSON_Success(t *testing.T) {
	Convey("ParseJSON decodes and validates a good payload", t, func() {
		body := `{"name":"ry","age":5}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

		out, err := ParseJSON[testIn](req)
		So(err, ShouldBeNil)
		So(out.Name, ShouldEqual, "ry")
		So(out.Age, ShouldEqual, 5)
	})
}

// TestParseJSON_Empty tests for an empty body and ensures error code
func TestParseJSON_Empty(t *testing.T) {
	Convey("ParseJSON returns ErrorCodeJSON on empty body", t, func() {
		req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		_, err := ParseJSON[testIn](req)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeJSON), ShouldBeTrue)
	})
}

// TestParseJSON_UnknownField handles testing unknown fields, which are default
func TestParseJSON_UnknownField(t *testing.T) {
	Convey("ParseJSON returns ErrorCodeJSON when unknown fields are present", t, func() {
		body := `{"name":"ry","age":5,"extra":true}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

		_, err := ParseJSON[testIn](req)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeJSON), ShouldBeTrue)
	})
}

// TestParseJSON_InvalidJSON ensures we handle malformed JSON
func TestParseJSON_InvalidJSON(t *testing.T) {
	Convey("ParseJSON returns ErrorCodeJSON on malformed JSON", t, func() {
		body := `{"name": "ry", "age": 5` // deliberate missing closing brace
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

		_, err := ParseJSON[testIn](req)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeJSON), ShouldBeTrue)
	})
}

// TestParseJSON_ValidationError makes sure there's context in error messages when handling validation
func TestParseJSON_ValidationError(t *testing.T) {
	Convey("ParseJSON returns ErrorCodeValidation and sets field", t, func() {
		body := `{"name":"r","age":0}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

		_, err := ParseJSON[testIn](req)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeValidation), ShouldBeTrue)

		var e *lumErrors.Error
		So(As(err, &e), ShouldBeTrue)
		So(e.Field(), ShouldNotBeBlank)
	})
}

// TestParseJSON_MaxBytes tests POST max bytes
func TestParseJSON_MaxBytes(t *testing.T) {
	Convey("ParseJSON enforces MaxBytes and returns ErrorCodeJSON when exceeded", t, func() {
		payload := `{"name":"really-long-name","age":9}`
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewBufferString(payload)))

		opts := JSONOptions{MaxBytes: 4, DisallowUnknown: true, AllowEmptyBody: false}
		_, err := ParseJSON[testIn](req, opts)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeJSON), ShouldBeTrue)
	})
}

// TestBindJSON tests binding JSON responses
func TestBindJSON(t *testing.T) {
	Convey("BindJSON parses/validates and stores payload for FromContext", t, func() {
		r := NewRouter(RouterOptions{WithLogger: false, WithRequestID: false, WithRecovery: false})

		// Attach per-route middleware via Group to avoid signature drift.
		r.Group(func(gr chi.Router) {
			gr.Use(BindJSON[testIn]())
			gr.Post("/faux", func(w http.ResponseWriter, req *http.Request) {
				val := FromContext[testIn](req)
				if val == nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if val.Name == "ry" && val.Age == 3 {
					w.WriteHeader(http.StatusCreated)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
		})

		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Post(ts.URL+"/faux", "application/json", bytes.NewBufferString(`{"name":"ry","age":3}`))
		So(err, ShouldBeNil)
		defer res.Body.Close()

		So(res.StatusCode, ShouldEqual, http.StatusCreated)
	})
}

// TestRegisterValidation_CustomTag tests custom validations
func TestRegisterValidation_CustomTag(t *testing.T) {
	Convey("Custom validation via RegisterValidation is applied", t, func() {
		So(RegisterValidation("eqfive", func(fl validator.FieldLevel) bool {
			return fl.Field().Int() == 5
		}), ShouldBeNil)

		type in struct {
			Age int `json:"age" validate:"eqfive"`
		}

		reqBad := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"age":4}`))
		_, err := ParseJSON[in](reqBad)
		So(err, ShouldNotBeNil)
		So(lumErrors.IsErrorCode(err, lumErrors.ErrorCodeValidation), ShouldBeTrue)

		reqOK := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"age":5}`))
		_, err = ParseJSON[in](reqOK)
		So(err, ShouldBeNil)
	})
}
