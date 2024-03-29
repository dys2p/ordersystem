package main

import (
	"fmt"
	"net/http"

	"github.com/dys2p/ordersystem/html"
	"github.com/julienschmidt/httprouter"
)

type HandlerErrFunc func(http.ResponseWriter, *http.Request) error

func auth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sessionManager.Exists(r.Context(), "username") {
			f(w, r)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}

func client(f HandlerErrFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			var msg string
			if err == ErrNotFound {
				msg = "Die Seite wurde nicht gefunden, die Aktion ist nicht erlaubt oder du bist nicht angemeldet."
			} else {
				msg = fmt.Sprintf("Interner Fehler: %s", err.Error())
			}
			html.ClientError.Execute(w, struct {
				AuthorizedCollID string
				Msg              string
			}{
				sessionCollID(r),
				msg,
			})
		}
	}
}

// requires authentication
func clientWithCollection(f func(w http.ResponseWriter, r *http.Request, coll *Collection) error) http.HandlerFunc {
	return client(
		func(w http.ResponseWriter, r *http.Request) error {
			var collID = httprouter.ParamsFromContext(r.Context()).ByName("collid")
			if len(collID) > 10 {
				collID = collID[:10]
			}
			if sessionManager.GetString(r.Context(), "coll-id") != collID {
				return ErrNotFound // not logged in, or into another collection
			}
			var coll, err = db.ReadColl(collID)
			if err != nil {
				return err
			}
			return f(w, r, coll)
		},
	)
}

func store(f HandlerErrFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			html.StoreError.Execute(w, err.Error())
		}
	}
}

func storeWithCollection(f func(w http.ResponseWriter, r *http.Request, coll *Collection) error) http.HandlerFunc {
	return store(
		func(w http.ResponseWriter, r *http.Request) error {
			var collID = httprouter.ParamsFromContext(r.Context()).ByName("collid")
			if len(collID) > 10 {
				collID = collID[:10]
			}
			var coll, err = db.ReadColl(collID)
			if err != nil {
				return err
			}
			return f(w, r, coll)
		},
	)
}

func storeWithTask(f func(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error) http.HandlerFunc {
	return storeWithCollection(
		func(w http.ResponseWriter, r *http.Request, coll *Collection) error {
			// select task from collection, don't get it from the database (then we had to verifiy that it belongs to the given collection)
			var taskID = httprouter.ParamsFromContext(r.Context()).ByName("taskid")
			for _, t := range coll.Tasks {
				if t.ID == taskID {
					return f(w, r, coll, t)
				}
			}
			return ErrNotFound
		},
	)
}
