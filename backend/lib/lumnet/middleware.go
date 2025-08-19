package lumnet

import (
	commonErrors "lumium/lib/errors"
	"lumium/lib/logger"

	"fmt"
	"net/http"
	"runtime/debug"
)

// sendRequestID sets the request ID into the header
func sendRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get(requestIDHeader) == "" {
			// middleware.RequestID must run before this; header is set by respond.go via RenderError too
			id := GetRequestID(r)
			if id != "" {
				w.Header().Set(requestIDHeader, id)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Recovery catches panics, logs a stack, and emits a standardized 500 JSON error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				l := logger.Get()
				lp := &l
				lp.Error().
					Str("stack", string(debug.Stack())).
					Msg(fmt.Sprintf("%v", rvr))
				RenderError(w, r, commonErrors.PanicErrf("internal server error"))
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
