package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"

	"lumium/lib/config"
	"lumium/lib/logger"
	"lumium/lib/lumnet"
	"lumium/lib/store"
	auth "lumium/services/api/auth"
	apihandlers "lumium/services/api/handlers"

	docs "lumium/services/api/docs"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

const serviceName = "lumium_api"

// WhoAmIResponse is the structure for declaring the instance
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
	// We can nuke, rename, or completely rewrite Serve under the hood,
	// and our tests don't even flinch
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

	mountRoutes(r, db)
	mountSwagger(r)

	return r
}

func mountRoutes(r *chi.Mux, db any) {
	if pool, ok := db.(*pgxpool.Pool); ok {
		r.Route("/api/v1", func(api chi.Router) {
			app := apihandlers.NewApp(pool)
			apihandlers.MountAPI(api,
				auth.New(app), // mounts /auth under /api/v1
			)
		})
	}
}

func mountSwagger(r *chi.Mux) {
	// Make “Try it out” use /api/v1 prefix
	docs.SwaggerInfoapi.BasePath = "/api/v1"

	// Serve a runtime-mutated OpenAPI JSON that injects examples from env.
	// MUST come before the /* UI handler so it isn't shadowed
	r.Get("/api/docs/doc.json", serveSwaggerJSONWithExamples())

	// Serve Swagger UI
	r.Handle("/api/docs/*", httpSwagger.Handler(httpSwagger.InstanceName("api")))
	r.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/docs/", http.StatusPermanentRedirect)
	})
}

// Run I don't even see the code... All I see is blonde, brunette, redhead...
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

// serveSwaggerJSONWithExamples reads the generated spec, injects runtime examples,
// and returns the modified JSON.
func serveSwaggerJSONWithExamples() http.HandlerFunc {
	exEmail := config.MayString("API_DOCS_DEFAULT_EMAIL", "user@lumiumapp.com")
	exPass := config.MayString("API_DOCS_DEFAULT_PASSWORD", "12345678")
	exTenant := config.MayString("API_DOCS_DEFAULT_TENANT_ID", "")

	raw := docs.SwaggerInfoapi.ReadDoc()

	return func(w http.ResponseWriter, r *http.Request) {
		var spec map[string]any
		if err := json.Unmarshal([]byte(raw), &spec); err != nil {
			http.Error(w, "spec parse error", http.StatusInternalServerError)
			return
		}

		loginExample := map[string]any{
			"email":    exEmail,
			"password": exPass,
		}
		if exTenant != "" {
			loginExample["tenant_id"] = exTenant
		}
		setSchemaExamples(spec, []string{"LoginDTO", "auth.LoginDTO"}, loginExample)
		setSchemaRequired(spec, []string{"LoginDTO", "auth.LoginDTO"}, []string{"email", "password"})
		for _, p := range []string{"/auth/login", "/api/v1/auth/login"} {
			setOperationRequestExample(spec, p, "post", loginExample)
		}

		signupExample := map[string]any{
			"email":    exEmail,
			"password": exPass,
		}
		if v := config.MayString("API_DOCS_DEFAULT_NAME", "Demo User"); v != "" {
			signupExample["name"] = v
		}
		if v := config.MayString("API_DOCS_DEFAULT_TENANT_SLUG", ""); v != "" {
			signupExample["tenant_slug"] = v
		}
		setSchemaExamples(spec, []string{"SignupDTO", "auth.SignupDTO"}, signupExample)
		setSchemaRequired(spec, []string{"SignupDTO", "auth.SignupDTO"}, []string{"email", "password"})
		for _, p := range []string{"/auth/register", "/api/v1/auth/register"} {
			setOperationRequestExample(spec, p, "post", signupExample)
		}

		forgotExample := map[string]any{
			"email": exEmail,
		}
		setSchemaExamples(spec, []string{"ForgotDTO", "auth.ForgotDTO"}, forgotExample)
		setSchemaRequired(spec, []string{"ForgotDTO", "auth.ForgotDTO"}, []string{"email"})
		for _, p := range []string{"/auth/forgot", "/api/v1/auth/forgot"} {
			setOperationRequestExample(spec, p, "post", forgotExample)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(spec)
	}
}

func setSchemaExamples(spec map[string]any, schemaKeys []string, example map[string]any) {
	// Swagger v2
	if defs, ok := spec["definitions"].(map[string]any); ok {
		for _, key := range schemaKeys {
			if s, ok := defs[key].(map[string]any); ok {
				s["example"] = example
				if props, ok := s["properties"].(map[string]any); ok {
					for k, v := range example {
						if p, ok := props[k].(map[string]any); ok {
							p["example"] = v
						}
					}
				}
			}
		}
	}
}

func setSchemaRequired(spec map[string]any, schemaKeys []string, required []string) {
	// Swagger v2
	if defs, ok := spec["definitions"].(map[string]any); ok {
		for _, key := range schemaKeys {
			if s, ok := defs[key].(map[string]any); ok {
				s["required"] = required
			}
		}
	}
}

func setOperationRequestExample(spec map[string]any, path, method string, example map[string]any) {
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		return
	}
	node, ok := paths[path].(map[string]any)
	if !ok {
		return
	}
	op, ok := node[strings.ToLower(method)].(map[string]any)
	if !ok {
		return
	}

	// Swagger v2: put example on body param schema
	if params, ok := op["parameters"].([]any); ok {
		for _, pi := range params {
			if p, ok := pi.(map[string]any); ok && p["in"] == "body" {
				if sch, ok := p["schema"].(map[string]any); ok {
					sch["example"] = example
				}
			}
		}
	}
}
