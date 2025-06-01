package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dys2p/eco/ssg"
	"github.com/dys2p/ordersystem"
	"github.com/dys2p/ordersystem/html"
	"github.com/julienschmidt/httprouter"
)

type HandlerErrFunc func(http.ResponseWriter, *http.Request) error

func (srv *Server) auth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if srv.Sessions.Exists(r.Context(), "username") {
			f(w, r)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}

func (srv *Server) client(f HandlerErrFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			var msg string
			if err == ErrNotFound {
				msg = "Die Seite wurde nicht gefunden, die Aktion ist nicht erlaubt oder du bist nicht angemeldet."
			} else {
				msg = fmt.Sprintf("Interner Fehler: %s", err.Error())
			}
			if err := html.ClientError.Execute(w, struct {
				html.TemplateData
				Msg string
			}{
				TemplateData: html.TemplateData{
					TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
					AuthorizedCollID: srv.sessionCollID(r),
				},
				Msg: msg,
			}); err != nil {
				log.Printf("error executing error template: %v", err)
			}
		}
	}
}

// requires authentication
func (srv *Server) clientWithCollection(f func(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error) http.HandlerFunc {
	return srv.client(
		func(w http.ResponseWriter, r *http.Request) error {
			var collID = httprouter.ParamsFromContext(r.Context()).ByName("collid")
			if len(collID) > 10 {
				collID = collID[:10]
			}
			if srv.Sessions.GetString(r.Context(), "coll-id") != collID {
				return ErrNotFound // not logged in, or into another collection
			}
			var coll, err = srv.DB.ReadColl(collID)
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

func (srv *Server) storeWithCollection(f func(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error) http.HandlerFunc {
	return store(
		func(w http.ResponseWriter, r *http.Request) error {
			var collID = httprouter.ParamsFromContext(r.Context()).ByName("collid")
			if len(collID) > 10 {
				collID = collID[:10]
			}
			var coll, err = srv.DB.ReadColl(collID)
			if err != nil {
				return err
			}
			return f(w, r, coll)
		},
	)
}

func (srv *Server) storeWithTask(f func(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error) http.HandlerFunc {
	return srv.storeWithCollection(
		func(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
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
