package lumnet

import "net/http"

// Reply is a deferred writer: call it to emit the response
type Reply func(http.ResponseWriter, *http.Request)

// Adapt turns a Reply-returning handler into a net/http handler
func Adapt(fn func(http.ResponseWriter, *http.Request) Reply) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rep := fn(w, r); rep != nil {
			rep(w, r)
		}
	}
}

// reply constructors wrapping our existing writers

func OKR(v any) Reply {
	return func(w http.ResponseWriter, r *http.Request) { OK(w, r, v) }
}
func CreatedR(v any, location string) Reply {
	return func(w http.ResponseWriter, r *http.Request) { Created(w, r, v, location) }
}
func NoContentR() Reply {
	return func(w http.ResponseWriter, r *http.Request) { NoContent(w, r) }
}
func ErrorR(err error) Reply {
	return func(w http.ResponseWriter, r *http.Request) { RenderError(w, r, err) }
}
func JSONStatusR(p map[string]any, status int) Reply {
	return func(w http.ResponseWriter, r *http.Request) { JSONStatus(w, r, p, status) }
}
