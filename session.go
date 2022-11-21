package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

var sessionManager *scs.SessionManager

func initSessionManager() error {

	sessionDB, err := sql.Open("sqlite3", "/var/lib/ordersystem/sessions.sqlite3?_busy_timeout=10000&_journal=WAL&_sync=NORMAL&cache=shared")
	if err != nil {
		return err
	}

	if _, err := sessionDB.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, data BLOB NOT NULL, expiry REAL NOT NULL);
		CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);
	`); err != nil {
		return err
	}

	sessionManager = scs.New()
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.Persist = false
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode // prevent CSRF
	sessionManager.Cookie.Secure = false
	sessionManager.IdleTimeout = 3 * time.Hour
	sessionManager.Lifetime = 60 * 24 * time.Hour // "absolute expiry which is set when the session is first created and does not change"
	sessionManager.Store = sqlite3store.New(sessionDB)
	return nil
}

func loginClient(ctx context.Context, collID string) {
	sessionManager.Put(ctx, "coll-id", collID) // "any existing value for the key will be replaced"
}

func loginStore(ctx context.Context, username string) {
	sessionManager.Put(ctx, "username", username)
}

func logout(ctx context.Context) {
	sessionManager.Destroy(ctx)
}

// adds a notification to the session
func notify(ctx context.Context, format string, a ...interface{}) {
	var ns, _ = sessionManager.Get(ctx, "ns").([]string)
	ns = append(ns, fmt.Sprintf(format, a...))
	sessionManager.Put(ctx, "ns", ns)
}

// returns and removes all notifications from the session
func notifications(ctx context.Context) []string {
	ns, _ := sessionManager.Pop(ctx, "ns").([]string)
	return ns
}

func sessionCollID(r *http.Request) string {
	return sessionManager.GetString(r.Context(), "coll-id")
}
