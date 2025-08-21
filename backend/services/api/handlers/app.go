// Package handlers is the convenience utility for definining controllers
package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Resource is anything that can wire routes onto a chi.Router
type Resource interface {
	Wire(r chi.Router)
}

// App holds shared deps for resources (start with DB, expand later if needed)
type App struct {
	DB *pgxpool.Pool
}

// NewApp accepts a database accessor & returns a new app
func NewApp(db *pgxpool.Pool) *App { return &App{DB: db} }

// MountAPI mounts one or more resources under the given router
func MountAPI(r chi.Router, resources ...Resource) {
	for _, res := range resources {
		res.Wire(r)
	}
}
