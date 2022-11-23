package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dchest/captcha"
)

func init() {

	db, err := sql.Open("sqlite3", filepath.Join(os.Getenv("STATE_DIRECTORY"), "captcha.sqlite3?_busy_timeout=10000&_journal=WAL&_sync=NORMAL&cache=shared"))
	if err != nil {
		panic(err)
	}
	store, err := NewDBStore(db)
	if err != nil {
		panic(err)
	}
	captcha.SetCustomStore(store)

	go func() {
		var ticker = time.NewTicker(1 * time.Hour)
		for range ticker.C {
			store.Cleanup()
		}
	}()
}

type DBStore struct {
	sqlDB *sql.DB
	clean *sql.Stmt
	del   *sql.Stmt
	get   *sql.Stmt
	set   *sql.Stmt
}

func NewDBStore(db *sql.DB) (*DBStore, error) {

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS captcha (
			id TEXT PRIMARY KEY,
			time INTEGER NOT NULL,
			digits BLOB NOT NULL
		)`); err != nil {
		return nil, err
	}

	var store = &DBStore{
		sqlDB: db,
	}
	var err error

	store.clean, err = db.Prepare("DELETE FROM captcha WHERE time < ?")
	if err != nil {
		return nil, err
	}

	store.del, err = db.Prepare("DELETE FROM captcha WHERE id = ?")
	if err != nil {
		return nil, err
	}

	store.get, err = db.Prepare("SELECT digits FROM captcha WHERE id = ? LIMIT 1")
	if err != nil {
		return nil, err
	}

	store.set, err = db.Prepare("INSERT OR REPLACE INTO captcha (id, time, digits) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Cleanup removes captchas older than two days.
func (s *DBStore) Cleanup() {
	if _, err := s.clean.Exec(time.Now().AddDate(0, 0, -2).Unix()); err != nil {
		log.Printf("captcha store cleanup: %v", err)
	}
}

// Set sets the digits for the captcha id.
func (s *DBStore) Set(id string, digits []byte) {
	if _, err := s.set.Exec(id, time.Now().Unix(), digits); err != nil {
		log.Printf("captcha store set: %v", err)
	}
}

// Get returns stored digits for the captcha id. Clear indicates
// whether the captcha must be deleted from the store.
func (s *DBStore) Get(id string, clear bool) (digits []byte) {
	if err := s.get.QueryRow(id).Scan(&digits); err != nil {
		if err == sql.ErrNoRows {
			// no problem, happens if a POST request is repeated
		} else {
			log.Printf("captcha store get: %v", err)
		}
	}
	if clear {
		if _, err := s.del.Exec(id); err != nil {
			log.Printf("captcha store delete: %v", err)
		}
	}
	return
}
