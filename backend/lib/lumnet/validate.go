package lumnet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	commonErrors "lumium/lib/errors"
	"lumium/lib/logger"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// ctx key for parsed JSON payload
type ctxKey uint8

const (
	bindJSONPayloadKey ctxKey = iota
)

type (
	// Aliases for convenience
	FieldLevel   = validator.FieldLevel
	UT           = ut.Translator
	FieldError   = validator.FieldError
	ValidatorSvc struct {
		Validator  *validator.Validate
		Translator ut.Translator
	}

	// JSONOptions control decoding behavior
	JSONOptions struct {
		MaxBytes        int64 // default 1MB
		DisallowUnknown bool  // default true
		AllowEmptyBody  bool  // default false
	}
)

var (
	vOnce sync.Once
	vSvc  *ValidatorSvc
)

// InitValidator initializes (or returns) the shared Validator service
func InitValidator() *ValidatorSvc {
	vOnce.Do(func() {
		enLoc := en.New()
		uni := ut.New(enLoc, enLoc)
		trans, _ := uni.GetTranslator("en")

		v := validator.New(validator.WithRequiredStructEnabled())

		// prefer JSON tag names in messages
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			tag := fld.Tag.Get("json")
			if tag == "-" || tag == "" {
				return fld.Name
			}
			if idx := strings.Index(tag, ","); idx >= 0 {
				tag = tag[:idx]
			}
			return tag
		})

		// base translations
		_ = en_translations.RegisterDefaultTranslations(v, trans)

		// concise overrides
		registerShortMin(v, trans)
		registerShortMax(v, trans)
		registerCommaInts(v, trans)

		vSvc = &ValidatorSvc{Validator: v, Translator: trans}
	})
	return vSvc
}

// GetValidator returns the singleton, or constructs it
func GetValidator() *ValidatorSvc {
	if vSvc == nil {
		return InitValidator()
	}
	return vSvc
}

// RegisterValidation adds a validation tag
func RegisterValidation(tag string, fn validator.Func) error {
	return GetValidator().Validator.RegisterValidation(tag, fn)
}

func defaultJSONOptions() JSONOptions {
	return JSONOptions{
		MaxBytes:        1 << 20, // 1MB
		DisallowUnknown: true,
		AllowEmptyBody:  false,
	}
}

// ParseJSON parses the body into T, validates, and returns domain errors on failure
func ParseJSON[T any](r *http.Request, opts ...JSONOptions) (T, error) {
	var zero T
	o := defaultJSONOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	defer r.Body.Close()

	var reader io.Reader

	// empty body guard with a peek
	if !o.AllowEmptyBody {
		buf := make([]byte, 1)
		n, _ := r.Body.Read(buf)
		if n == 0 {
			return zero, commonErrors.JSONErrf("empty body")
		}
		combined := io.MultiReader(bytes.NewReader(buf[:n]), r.Body)

		if o.MaxBytes > 0 {
			reader = io.LimitReader(combined, o.MaxBytes)
		} else {
			reader = combined
		}
	} else {
		// no peek path
		if o.MaxBytes > 0 {
			reader = io.LimitReader(r.Body, o.MaxBytes)
		} else {
			reader = r.Body
		}
	}

	dec := json.NewDecoder(reader)
	if o.DisallowUnknown {
		dec.DisallowUnknownFields()
	}

	var dst T
	if err := dec.Decode(&dst); err != nil {
		return zero, commonErrors.JSONErrf("invalid JSON: %v", err)
	}

	// no trailing junk
	if dec.More() {
		return zero, commonErrors.JSONErrf("unexpected trailing data")
	}

	// validate
	if err := GetValidator().Validator.Struct(dst); err != nil {
		if inv, ok := err.(*validator.InvalidValidationError); ok {
			l := logger.Get()
			lp := &l
			lp.Error().Err(inv).Msg("validator internal error")
			return zero, commonErrors.JSONErrf("validation error")
		}
		field, msg := ValidationFieldAndMessage(err)
		return zero, commonErrors.NewValidationError(commonErrors.ErrorCodeValidation, msg, field)
	}

	return dst, nil
}

// BindJSON parses + validates into context under bindJSONPayloadKey
func BindJSON[T any](opts ...JSONOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val, err := ParseJSON[T](r, opts...)
			if err != nil {
				RenderError(w, r, err)
				return
			}
			ctx := context.WithValue(r.Context(), bindJSONPayloadKey, &val)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext extracts parsed payload
func FromContext[T any](r *http.Request) *T {
	v, _ := r.Context().Value(bindJSONPayloadKey).(*T)
	return v
}

// ValidationFieldAndMessage returns the first invalid field and a concise, translated message.
func ValidationFieldAndMessage(err error) (field, message string) {
	if err == nil {
		return "", ""
	}
	if inv, ok := err.(*validator.InvalidValidationError); ok {
		return "", inv.Error()
	}
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			// Field name already prefers JSON tag via RegisterTagNameFunc
			return fe.Field(), fe.Translate(GetValidator().Translator)
		}
	}
	return "", err.Error()
}

// GetRequestID helps with convenience to get request id safely
func GetRequestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}

// stdlib errors.As alias to keep imports tidy in this package
func As(err error, target any) bool { return errors.As(err, target) }

func registerShortMin(v *validator.Validate, trans ut.Translator) {
	_ = v.RegisterTranslation("min", trans,
		func(ut ut.Translator) error {
			return ut.Add("min", "{0} must be \u2265 {1}", true) // ≥
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			msg, _ := ut.T("min", fe.Field(), fe.Param())
			return msg
		},
	)
}

func registerShortMax(v *validator.Validate, trans ut.Translator) {
	_ = v.RegisterTranslation("max", trans,
		func(ut ut.Translator) error {
			return ut.Add("max", "{0} must be \u2264 {1}", true) // ≤
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			msg, _ := ut.T("max", fe.Field(), fe.Param())
			return msg
		},
	)
}

// comma_ints -> "field must be a comma-separated list of integers"
func registerCommaInts(v *validator.Validate, trans ut.Translator) {
	_ = v.RegisterTranslation("comma_ints", trans,
		func(ut ut.Translator) error {
			return ut.Add("comma_ints", "{0} must be a comma-separated list of integers", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			msg, _ := ut.T("comma_ints", fe.Field())
			return msg
		},
	)
}
