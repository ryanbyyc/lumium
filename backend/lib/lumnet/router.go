package lumnet

import (
	lumErrors "lumium/lib/errors"
	"lumium/lib/logger"

	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

// requestIDHeader is our header
const requestIDHeader = "X-Request-Id"

// ErrResponse is the JSON error envelope used by RenderError
type ErrResponse struct {
	HTTPStatusCode       int             `json:"status_code"`
	StatusText           string          `json:"status"`
	AppCode              int64           `json:"code"`
	ErrorText            string          `json:"error"`
	ValidationErrorField string          `json:"validation_field,omitempty"`
	RequestID            string          `json:"request_id,omitempty"`
	Payload              *lumErrors.Wire `json:"payload,omitempty"`
}

// RouterOptions toggles middleware
type RouterOptions struct {
	WithLogger    bool
	WithRequestID bool
	WithRecovery  bool
	WithCORS      bool
	CORS          *cors.Options
}

// NewRouter builds a chi router with defaults and JSON content type
func NewRouter(opts RouterOptions) *chi.Mux {
	r := chi.NewRouter()
	InitValidator()

	r.Use(render.SetContentType(render.ContentTypeJSON))

	if opts.WithCORS {
		co := opts.CORS
		if co == nil {
			co = defaultCORS()
		}
		r.Use(cors.Handler(*co))
	}

	if opts.WithLogger {
		r.Use(middleware.Logger)
	}
	if opts.WithRequestID {
		r.Use(middleware.RequestID)
		r.Use(sendRequestID)
	}
	if opts.WithRecovery {
		r.Use(Recovery)
	}
	return r
}

func defaultCORS() *cors.Options {
	// Allow common local dev UIs and an env override FRONTEND_ORIGIN (comma-separated)
	origins := []string{
		"http://localhost:9080",
		"http://127.0.0.1:9080",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
	if v := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")); v != "" {
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
	}

	return &cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Request-Id"},
		AllowCredentials: true, // flip to false if you don’t need cookies/auth
		MaxAge:           300,  // seconds
	}
}

// Serve runs the HTTP server on the given listener with graceful shutdown
// The graceful shutdown is important when clustering
func Serve(listener net.Listener, mux http.Handler, serviceName string) {
	l := logger.Get()

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	l.Info().Str("addr", listener.Addr().String()).Msgf("Starting %s service", serviceName)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(listener) }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		l.Info().Str("signal", sig.String()).Msg("Shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error().Err(err).Msg("HTTP server error")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		l.Error().Err(err).Msg("Graceful shutdown timed out; forcing close")
		_ = srv.Close()
	}
	l.Info().Msg("Server offline")
}
