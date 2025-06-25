package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

func initSessionManager() (*scs.SessionManager, error) {

	sessionDB, err := sql.Open("sqlite3", filepath.Join(os.Getenv("STATE_DIRECTORY"), "sessions.sqlite3?_busy_timeout=10000&_journal=WAL&_sync=NORMAL&cache=shared"))
	if err != nil {
		return nil, err
	}

	if _, err := sessionDB.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, data BLOB NOT NULL, expiry REAL NOT NULL);
		CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);
	`); err != nil {
		return nil, err
	}

	var sessionManager = scs.New()
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.Persist = false
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode // prevent CSRF
	sessionManager.Cookie.Secure = false
	sessionManager.IdleTimeout = 3 * 24 * time.Hour
	sessionManager.Lifetime = 60 * 24 * time.Hour // "absolute expiry which is set when the session is first created and does not change"
	sessionManager.Store = sqlite3store.New(sessionDB)
	return sessionManager, nil
}

func (srv *Server) loginClient(ctx context.Context, collID string) {
	srv.Sessions.Put(ctx, "coll-id", collID) // "any existing value for the key will be replaced"
}

func (srv *Server) loginStore(ctx context.Context, username string) {
	srv.Sessions.Put(ctx, "username", username)
}

func (srv *Server) logout(ctx context.Context) {
	srv.Sessions.Destroy(ctx)
}

// adds a notification to the session
func (srv *Server) notify(ctx context.Context, format string, a ...interface{}) {
	var ns, _ = srv.Sessions.Get(ctx, "ns").([]string)
	ns = append(ns, fmt.Sprintf(format, a...))
	srv.Sessions.Put(ctx, "ns", ns)
}

// returns and removes all notifications from the session
func (srv *Server) notifications(ctx context.Context) []string {
	ns, _ := srv.Sessions.Pop(ctx, "ns").([]string)
	return ns
}

func (srv *Server) sessionCollID(r *http.Request) string {
	return srv.Sessions.GetString(r.Context(), "coll-id")
}
