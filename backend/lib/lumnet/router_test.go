package lumnet

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestNewRouter assumes Request ID & JSON
func TestNewRouter(t *testing.T) {
	Convey("NewRouter sets JSON content type and adds X-Request-Id when enabled", t, func() {
		r := NewRouter(RouterOptions{
			WithLogger:    false, // keep test output quiet
			WithRequestID: true,
			WithRecovery:  false,
		})

		r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
			// write a JSON body so Content-Type is observable
			OK(w, req, map[string]any{"ok": true})
		})

		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Get(ts.URL + "/ping")
		So(err, ShouldBeNil)
		defer res.Body.Close()

		So(res.StatusCode, ShouldEqual, http.StatusOK)

		// Middleware should set JSON content type (render helpers would too)
		So(res.Header.Get("Content-Type"), ShouldContainSubstring, "application/json")

		// RequestID middleware + sendRequestID should surface header
		So(res.Header.Get(requestIDHeader), ShouldNotBeBlank)
	})
}

// TestNewRouter_RequestIDDisabled tests when request ID is disabled
func TestNewRouter_RequestIDDisabled(t *testing.T) {
	Convey("NewRouter does not set X-Request-Id when WithRequestID is false", t, func() {
		r := NewRouter(RouterOptions{
			WithLogger:    false,
			WithRequestID: false,
			WithRecovery:  false,
		})

		r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
			OK(w, req, map[string]any{"ok": true})
		})

		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Get(ts.URL + "/ping")
		So(err, ShouldBeNil)
		defer res.Body.Close()

		So(res.StatusCode, ShouldEqual, http.StatusOK)
		So(res.Header.Get(requestIDHeader), ShouldBeBlank)
	})
}

// testListener forces http Serve to exit immediately by returning an Accept error
type testListener struct{}

// Accept is a shim for the net.Listener interface, used in tests via a seam to force an accept error
func (testListener) Accept() (net.Conn, error) { return nil, errors.New("accept error") }

// Close is a shim for the net.Listener interface, used in tests via a seam to close it
func (testListener) Close() error { return nil }

// Addr returns a dummy address for the test listener
func (testListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0} }

// TestServe_ListenerError returns when accept fails
func TestServe_ListenerError(t *testing.T) {
	Convey("Serve returns when listener.Accept fails", t, func() {
		done := make(chan struct{})

		// Create a Minimal mux
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Real world, Serve() runs forever (or until the listener errors/gets closed)
		// Wrap it in a gofunc so it starts in the background and is not blocking
		go func() {
			Serve(testListener{}, mux, "test-svc")
			close(done)
		}()

		select {
		case <-done:
			// ok
		case <-time.After(2 * time.Second):
			t.Fatal("Serve did not return on listener error in time")
		}
	})
}

// TestServe_GracefulOnSIGTERM ensures graceful shutdowns
func TestServe_GracefulOnSIGTERM(t *testing.T) {
	Convey("Serve shuts down gracefully on SIGTERM", t, func() {

		// TCP listener on an ephemeral port
		ln, err := net.Listen("tcp", "127.0.0.1:0")

		So(err, ShouldBeNil)
		defer ln.Close()

		mux := http.NewServeMux()
		mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		done := make(chan struct{})
		go func() {
			Serve(ln, mux, "test-svc")
			close(done)
		}()

		// Give the server a moment to start
		time.Sleep(1 * time.Second)

		// Send SIGTERM to this process; Serve listens for it and should shut down
		So(syscall.Kill(os.Getpid(), syscall.SIGTERM), ShouldBeNil)

		select {
		case <-done:

		// Making sure we shutdown after the termination signal
		case <-time.After(3 * time.Second):
			t.Fatal("Serve did not shut down after SIGTERM")
		}
	})
}
