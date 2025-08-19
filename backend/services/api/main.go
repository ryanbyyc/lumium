package main

import (
	"lumium/lib/config"
	"lumium/lib/logger"
	"lumium/lib/lumnet"
	"lumium/lib/store"

	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
)

// serviceName allows us to identify the service
const serviceName = "lumium_api"

type WhoAmIResponse struct {
	Hostname string `json:"hostname"`
	PID      int    `json:"pid"`
}

// closer is an interface that defines the structure for closing the db handler
// this is designed to support testing
type closer interface{ Close() }

type dbPinger interface {
	Ping(ctx context.Context) error
}

// variables for seaming in tests
var (
	listenFn = net.Listen

	// The beauty of keeping initHTTPFn is that its an injectable seam
	// We can nuke, rename, or completely rewrite Serve under the hood, and our tests don't even flinch
	initHTTPFn = lumnet.Serve

	// newHandlerFn return a small interface so tests can stub without importing store types
	newHandlerFn = func(initializePgx bool) closer { return store.NewHandler(initializePgx) }

	extractPingerFn = func(h closer) dbPinger {
		// real path: use the store.Handler's pgx pool
		return h.(*store.Handler).PgxPool
	}
)

// BuildRouter centralizes all route wiring
func BuildRouter(db dbPinger) http.Handler {
	r := lumnet.NewRouter(lumnet.RouterOptions{
		WithLogger:    true,
		WithRequestID: true,
		WithRecovery:  true,
		WithCORS:      true,
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	})

	// Endpoint to test readiness and avoids serving traffic when the DB is down
	r.Get("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if db == nil {
			http.Error(w, "database not initialized", http.StatusServiceUnavailable)
			return
		}

		// Basic ping / health check
		if err := db.Ping(context.Background()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`rdy`))
	})

	// Used during swarm testing to ensure different machineIDs
	// example:
	// for i in {1..10}; do curl -s http://localhost:9001/whoami; echo; done
	r.Get("/whoami", func(w http.ResponseWriter, _ *http.Request) {
		hostname, _ := os.Hostname()
		resp := WhoAmIResponse{Hostname: hostname, PID: os.Getpid()}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// if pool, ok := db.(*pgxpool.Pool); ok {
	// 	// r.Route("/api/v1", func(api chi.Router) {
	// 	// 	app := handlers.NewApp(pool)
	// 	// 	handlers.MountAPI(api,
	// 	// 		handlers.NewAuth(app),
	// 	// 	)
	// 	// })
	// }

	return r
}

// I don't even see the code... All I see is blonde, brunette, redhead...
func Run() {
	l := logger.Get()
	addr := config.MustPort("CORE_API_PORT")

	listener, err := listenFn("tcp", addr)
	if err != nil {
		// In TestRun_PortUnavailable we stub listenFn to return an error
		// so we can cover the early exit in Run()
		// This warning is expected in goConvey
		l.Warn().
			Str("addr", addr).
			Err(err).
			Msg("Port unavailable; exiting without starting server")
		return
	}

	// call newHandlerFn with true to create a postgres connection
	// We also defer the close
	h := newHandlerFn(true)
	defer h.Close()

	// Build the router
	r := BuildRouter(extractPingerFn(h))

	// Finally, init http with the service name and router
	initHTTPFn(listener, r, serviceName)
}

func main() {
	Run()
}
