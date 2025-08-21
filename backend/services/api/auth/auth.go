// Package auth implements the role of system authentication
package auth

import (
	"lumium/lib/lumnet"
	"lumium/lib/svckit"
	"lumium/services/api/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// svc embeds the shared Kit so we get DB/Repo/Cfg without redefining fields
type svc struct {
	*svckit.Kit[*pgxpool.Pool, Repo, Config]
}

// NewService defaults to NewRepo(), but can be overridden with WithRepo(...)
func NewService(db *pgxpool.Pool, c Config, o ...svckit.Opt[*pgxpool.Pool, Repo, Config]) Service {
	return &svc{Kit: svckit.New(db, NewRepo, c, o...)}
}

// Config returns the auth config
func (s *svc) Config() Config {
	return s.Cfg
}

// Auth is the wrapper for the /auth service
type Auth struct {
	app *handlers.App
	svc Service
}

type repo struct{}

// NewRepo creates a repo pointer
func NewRepo() Repo { return &repo{} }

// New creates a new Auth pointer
func New(app *handlers.App) *Auth {
	cfg := LoadConfig()
	svc := NewService(app.DB, cfg)
	return &Auth{app: app, svc: svc}
}

// Wire defines the HTTP endpoint structure
func (h *Auth) Wire(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", lumnet.Adapt(h.Login))
		r.Post("/register", lumnet.Adapt(h.Register))
		r.Post("/refresh", lumnet.Adapt(h.Refresh))
		r.Post("/logout", lumnet.Adapt(h.Logout))
		r.Get("/me", lumnet.Adapt(h.Me))

		r.Post("/mfa/challenge", lumnet.Adapt(h.MFAChallenge)) // optional resend/new
		r.Post("/mfa/verify", lumnet.Adapt(h.MFAVerify))

		r.Post("/forgot", lumnet.Adapt(h.Forgot)) // 202 always
		r.Post("/reset", lumnet.Adapt(h.Reset))   // { token, password }
	})
	lumnet.InitValidator()
}
