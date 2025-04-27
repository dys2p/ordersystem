package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dys2p/bitpay"
	"github.com/dys2p/btcpay"
	"github.com/dys2p/digitalgoods/userdb"
	"github.com/dys2p/eco/captcha"
	"github.com/dys2p/eco/diceware"
	"github.com/dys2p/eco/httputil"
	"github.com/dys2p/eco/id"
	"github.com/dys2p/eco/lang"
	"github.com/dys2p/eco/ssg"
	"github.com/dys2p/ordersystem"
	"github.com/dys2p/ordersystem/html"
	"github.com/dys2p/ordersystem/html/scripts"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("not found")

const MaxDiscountCents = 10

func main() {
	langs := lang.MakeLanguages(nil, "de") // after catalog.go is loaded

	log.SetFlags(0)

	// os flags

	var test = flag.Bool("test", false, "use btcpay dummy store")
	flag.Parse()

	// websites

	siteFiles, err := fs.Sub(html.Files, "order.proxysto.re")
	if err != nil {
		log.Fatalf("error opening site dir: %v", err)
	}
	staticFiles, err := fs.Sub(html.Files, "order.proxysto.re/static") // for staff router
	if err != nil {
		log.Fatalf("error opening static dir: %v", err)
	}

	staticSites, err := ssg.MakeWebsite(siteFiles, html.ClientSite, langs)
	if err != nil {
		log.Fatalf("error making static sites: %v", err)
	}

	// SQL db

	sqlDB, err := sql.Open("sqlite3", filepath.Join(os.Getenv("STATE_DIRECTORY"), "ordersystem.sqlite3?_busy_timeout=10000&_journal=WAL&_sync=NORMAL&cache=shared"))
	if err != nil {
		log.Printf("error opening database: %v", err)
		return
	}

	// captcha

	captcha.Initialize(filepath.Join(os.Getenv("STATE_DIRECTORY"), "captcha.sqlite3"))

	// userdb

	users, err := userdb.Open(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "users.json"))
	if err != nil {
		log.Printf("error opening userdb: %v", err)
		return
	}

	// bitpay

	bitpayClient, err := bitpay.LoadClient(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "bitpay.json"))
	if err != nil {
		log.Printf("error creating bitpay API client: %v", err)
		if !*test {
			return
		}
	}

	log.Printf("please make sure that your BTCPay store is paired to the public key (hex SIN): %s", bitpayClient.SINHex())

	// btcpay

	var btcpayStore btcpay.Store
	if *test {
		btcpayStore = btcpay.NewDummyStore()
		log.Println("\033[33m" + "warning: using btcpay dummy store" + "\033[0m")
	} else {
		btcpayStore, err = btcpay.Load(filepath.Join(os.Getenv("CONFIGURATION_DIRECTORY"), "btcpay.json"))
		if err != nil {
			log.Printf("error loading btcpay store: %v", err)
			return
		}

		log.Println("don't forget to set up the webhook for your store: /rpc")
		log.Println(`  Event: "A new payment has been received"`)
		log.Println(`  Event: "An invoice has been settled"`)
	}

	// session db

	sessions, err := initSessionManager()
	if err != nil {
		log.Printf("error initializing session manager: %v", err)
		return
	}

	// db

	db, err := ordersystem.NewDB(sqlDB)
	if err != nil {
		log.Printf("error creating database: %v", err)
		return
	}

	// server

	srv := &Server{
		BitpayClient: bitpayClient,
		BtcPayStore:  btcpayStore,
		DB:           db,
		Langs:        langs,
		Sessions:     sessions,
		Users:        users,
	}

	// bot

	var ticker = time.NewTicker(12 * time.Hour)
	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-ticker.C:
				wg.Add(1)
				srv.Bot()
				wg.Done()
			case id := <-botCollIDs:
				wg.Add(1)
				srv.BotColl(id)
				wg.Done()
			}
		}
	}()

	srv.Bot() // run now

	// http handlers

	var stop = make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var clientRouter = httprouter.New()
	clientRouter.ServeFiles("/static/*filepath", http.FS(httputil.ModTimeFS{staticFiles, time.Now()}))
	clientRouter.HandlerFunc(http.MethodGet, "/", srv.client(srv.clientHelloGet))
	clientRouter.HandlerFunc(http.MethodGet, "/create", srv.client(srv.clientCreateGet))
	clientRouter.HandlerFunc(http.MethodPost, "/create", srv.client(srv.clientCreatePost))
	clientRouter.HandlerFunc(http.MethodGet, "/current", srv.client(srv.clientCollCurrentGet))
	clientRouter.HandlerFunc(http.MethodGet, "/collection", srv.client(srv.clientCollLoginGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection", srv.client(srv.clientCollLoginPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid", srv.clientWithCollection(srv.clientCollViewGet))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/cancel", srv.clientWithCollection(srv.clientCollCancelGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/cancel", srv.clientWithCollection(srv.clientCollCancelPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/delete", srv.clientWithCollection(srv.clientCollDeleteGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/delete", srv.clientWithCollection(srv.clientCollDeletePost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/edit", srv.clientWithCollection(srv.clientCollEditGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/edit", srv.clientWithCollection(srv.clientCollEditPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/message", srv.clientWithCollection(srv.clientCollMessageGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/message", srv.clientWithCollection(srv.clientCollMessagePost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/pay-btcpay", srv.clientWithCollection(srv.clientCollPayBTCPayGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/pay-btcpay", srv.clientWithCollection(srv.clientCollPayBTCPayPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/submit", srv.clientWithCollection(srv.clientCollSubmitGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/submit", srv.clientWithCollection(srv.clientCollSubmitPost))

	clientRouter.HandlerFunc(http.MethodPost, "/rpc", srv.rpc)
	clientRouter.HandlerFunc(http.MethodGet, "/state", srv.client(srv.clientStateGet))
	clientRouter.HandlerFunc(http.MethodPost, "/state", srv.client(srv.clientStatePost))
	clientRouter.HandlerFunc(http.MethodPost, "/logout", srv.client(srv.clientLogoutPost))
	clientRouter.Handler("GET", "/captcha/:fn", captcha.Handler())
	clientRouter.ServeFiles("/scripts/*filepath", http.FS(scripts.Files))
	clientRouter.NotFound = staticSites.Handler(func(r *http.Request, td ssg.TemplateData) any {
		return html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		}
	}, langs.RedirectHandler())

	shutdownClientSrv := httputil.ListenAndServe("127.0.0.1:9000", srv.Sessions.LoadAndSave(clientRouter), stop)
	defer shutdownClientSrv()

	var storeRouter = httprouter.New()
	storeRouter.ServeFiles("/static/*filepath", http.FS(httputil.ModTimeFS{staticFiles, time.Now()}))
	storeRouter.HandlerFunc(http.MethodGet, "/login", store(srv.storeLoginGet))
	storeRouter.HandlerFunc(http.MethodPost, "/login", store(srv.storeLoginPost))
	// with authentication:
	storeRouter.HandlerFunc(http.MethodGet, "/", srv.auth(store(srv.storeIndexGet)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid", srv.auth(srv.storeWithCollection(srv.storeCollViewGet)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/accept", srv.auth(srv.storeWithCollection(srv.storeCollAcceptGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/accept", srv.auth(srv.storeWithCollection(srv.storeCollAcceptPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-payment", srv.auth(srv.storeWithCollection(srv.storeCollConfirmPaymentGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-payment", srv.auth(srv.storeWithCollection(srv.storeCollConfirmPaymentPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-pickup", srv.auth(srv.storeWithCollection(srv.storeCollConfirmPickupGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-pickup", srv.auth(srv.storeWithCollection(srv.storeCollConfirmPickupPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-pickup/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmPickupGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-pickup/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmPickupPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-reshipped", srv.auth(srv.storeWithCollection(srv.storeCollConfirmReshippedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-reshipped", srv.auth(srv.storeWithCollection(srv.storeCollConfirmReshippedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-reshipped/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmReshippedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-reshipped/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmReshippedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/delete", srv.auth(srv.storeWithCollection(srv.storeCollDeleteGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/delete", srv.auth(srv.storeWithCollection(srv.storeCollDeletePost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/edit", srv.auth(srv.storeWithCollection(srv.storeCollEditGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/edit", srv.auth(srv.storeWithCollection(srv.storeCollEditPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/mark-spam", srv.auth(srv.storeWithCollection(srv.storeCollMarkSpamGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/mark-spam", srv.auth(srv.storeWithCollection(srv.storeCollMarkSpamPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/message", srv.auth(srv.storeWithCollection(srv.storeCollMessageGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/message", srv.auth(srv.storeWithCollection(srv.storeCollMessagePost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/price-rised", srv.auth(srv.storeWithCollection(srv.storeCollPriceRisedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/price-rised", srv.auth(srv.storeWithCollection(srv.storeCollPriceRisedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/return", srv.auth(srv.storeWithCollection(srv.storeCollReturnGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/return", srv.auth(srv.storeWithCollection(srv.storeCollReturnPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/reject", srv.auth(srv.storeWithCollection(srv.storeCollRejectGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/reject", srv.auth(srv.storeWithCollection(srv.storeCollRejectPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/submit", srv.auth(srv.storeWithCollection(srv.storeCollSubmitGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/submit", srv.auth(srv.storeWithCollection(srv.storeCollSubmitPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-arrived/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmArrivedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-arrived/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmArrivedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-ordered/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmOrderedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-ordered/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskConfirmOrderedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/mark-failed/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskMarkFailedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/mark-failed/:taskid", srv.auth(srv.storeWithTask(srv.storeTaskMarkFailedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/export", srv.auth(store(srv.storeExport)))
	storeRouter.HandlerFunc(http.MethodPost, "/logout", store(srv.storeLogoutPost))
	storeRouter.ServeFiles("/scripts/*filepath", http.FS(scripts.Files))

	shutdownStoreSrv := httputil.ListenAndServe("127.0.0.1:9001", srv.Sessions.LoadAndSave(storeRouter), stop)
	defer shutdownStoreSrv()

	log.Printf("listening to 127.0.0.1:9000 and 127.0.0.1:9001")
	<-stop // blocks
	log.Println("shutting down")
	wg.Wait()
}

type collView struct {
	html.TemplateData
	*ordersystem.Collection
	Actor         ordersystem.Actor
	ReadOnly      bool
	ShowHints     bool
	Notifications []string
}

type clientHello struct {
	html.TemplateData
	Notifications []string
}

func (srv *Server) clientHelloGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientHello.Execute(w, clientHello{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Notifications: srv.notifications(r.Context()),
	})
}

type clientCreate struct {
	html.TemplateData
	CollID              string
	CollIDErr           bool
	CollPass            string
	CollPassErr         bool
	Captcha             captcha.TemplateData
	CheckWrittenDown    bool
	CheckWrittenDownErr bool
}

func isID(s string) bool {
	if len(s) != 6 && len(s) != 10 {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune(id.AlphanumCaseInsensitiveDigits, r) {
			return false
		}
	}
	return true
}

func (data *clientCreate) Valid() bool {
	if !isID(data.CollID) {
		data.CollIDErr = true
	}
	if data.CollPass == "" {
		data.CollPassErr = true
	}
	if !data.CheckWrittenDown {
		data.CheckWrittenDownErr = true
	}
	return !data.CollIDErr && !data.CollPassErr && !data.Captcha.Err && !data.CheckWrittenDownErr
}

func (srv *Server) clientCreateGet(w http.ResponseWriter, r *http.Request) error {
	collPass, err := diceware.Length(5, diceware.German)
	if err != nil {
		return err
	}
	return html.ClientCreate.Execute(w, &clientCreate{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		CollID:   id.New(6, id.AlphanumCaseInsensitiveDigits),
		CollPass: collPass,
		Captcha: captcha.TemplateData{
			ID: captcha.New(),
		},
	})
}

func (srv *Server) clientCreatePost(w http.ResponseWriter, r *http.Request) error {

	var data = &clientCreate{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		CollID:           strings.TrimSpace(r.PostFormValue("collection-id")),
		CollPass:         strings.TrimSpace(r.PostFormValue("collection-passphrase")),
		CheckWrittenDown: r.PostFormValue("check-written-down") != "",
	}

	if !captcha.Verify(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-answer")) {
		data.Captcha.Err = true
	}

	if !data.Valid() {
		data.Captcha.ID = captcha.New() // old captcha has been deleted during verification, so let's create a new one
		return html.ClientCreate.Execute(w, data)
	}

	var coll = &ordersystem.Collection{
		ID:    data.CollID,
		State: ordersystem.Draft,
	}

	if hash, err := ordersystem.HashPassword(data.CollPass); err == nil {
		coll.Pass = string(hash)
	} else {
		return err
	}

	if err := srv.DB.CreateCollection(coll); err != nil {
		return err
	}

	srv.loginClient(r.Context(), coll.ID)
	http.Redirect(w, r, fmt.Sprintf("/collection/%s/edit", coll.ID), http.StatusSeeOther)
	return nil
}

func (srv *Server) clientCollEditGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("edit") {
		return ErrNotFound
	}
	return html.ClientCollEdit.Execute(w, collView{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Actor:      ordersystem.Client,
		Collection: coll,
		ShowHints:  true,
	})
}

func (srv *Server) clientCollEditPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("edit") {
		return ErrNotFound
	}
	if err := coll.MergeJSON(ordersystem.Client, r.PostFormValue("data")); err != nil {
		return err
	}
	if err := srv.DB.UpdateCollAndTasks(coll); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type clientLogin struct {
	html.TemplateData
	CollID      string
	CollIDErr   bool
	CollPassErr bool
}

// success and error url for payment providers, so we don't reveal the collection ID to them
func (srv *Server) clientCollCurrentGet(w http.ResponseWriter, r *http.Request) error {
	if collID := srv.sessionCollID(r); collID != "" {
		http.Redirect(w, r, fmt.Sprintf("/collection/%s", collID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	return nil
}

func (srv *Server) clientCollLoginGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientCollLogin.Execute(w, &clientLogin{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
	})
}

func (srv *Server) clientCollLoginPost(w http.ResponseWriter, r *http.Request) error {
	var id = strings.TrimSpace(r.FormValue("collection-id"))
	var data = &clientLogin{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		CollID: id,
	}
	if !isID(id) {
		data.CollIDErr = true
		return html.ClientCollLogin.Execute(w, data)
	}
	var pass = strings.TrimSpace(r.PostFormValue("collection-passphrase"))
	var coll, err = srv.DB.ReadCollPass(id, pass)
	if err != nil {
		data.CollPassErr = true
		return html.ClientCollLogin.Execute(w, data)
	}

	srv.loginClient(r.Context(), coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) clientCollViewGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	return html.ClientCollView.Execute(w, collView{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Actor:         ordersystem.Client,
		Collection:    coll,
		ReadOnly:      true,
		Notifications: srv.notifications(r.Context()),
	})
}

type clientCollCancel struct {
	html.TemplateData
	*ordersystem.Collection
	Err bool
}

func (srv *Server) clientCollCancelGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("cancel") {
		return ErrNotFound
	}
	return html.ClientCollCancel.Execute(w, &clientCollCancel{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Collection: coll,
	})
}

func (srv *Server) clientCollCancelPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("cancel") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-cancel") == "" {
		return html.ClientCollCancel.Execute(w, &clientCollCancel{
			TemplateData: html.TemplateData{
				TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
				AuthorizedCollID: srv.sessionCollID(r),
			},
			Collection: coll,
			Err:        true,
		})
	}
	if err := srv.DB.UpdateCollState(ordersystem.Client, coll, ordersystem.Cancelled, 0, ""); err != nil {
		return err
	}

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

// used by client and store
type collDelete struct {
	html.TemplateData
	*ordersystem.Collection
	Err bool
}

func (srv *Server) clientCollDeleteGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("delete") {
		return ErrNotFound
	}
	return html.ClientCollDelete.Execute(w, &collDelete{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Collection: coll,
	})
}

func (srv *Server) clientCollDeletePost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("delete") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-delete") == "" {
		return html.ClientCollDelete.Execute(w, &collDelete{
			TemplateData: html.TemplateData{
				TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
				AuthorizedCollID: srv.sessionCollID(r),
			},
			Collection: coll,
			Err:        true,
		})
	}
	if err := srv.DB.Delete(ordersystem.Client, coll); err != nil {
		return err
	}

	srv.logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

type clientCollPayBTCPay struct {
	html.TemplateData
	*ordersystem.Collection
	Captcha captcha.TemplateData
}

func (srv *Server) clientCollPayBTCPayGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("pay") {
		return ErrNotFound
	}
	return html.ClientCollPayBTCPay.Execute(w, &clientCollPayBTCPay{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Collection: coll,
		Captcha: captcha.TemplateData{
			ID: captcha.New(),
		},
	})
}

func (srv *Server) clientCollPayBTCPayPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {

	if !coll.ClientCan("pay") {
		return ErrNotFound
	}
	if !captcha.Verify(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-answer")) {
		return html.ClientCollPayBTCPay.Execute(w, &clientCollPayBTCPay{
			TemplateData: html.TemplateData{
				TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
				AuthorizedCollID: srv.sessionCollID(r),
			},
			Collection: coll,
			Captcha: captcha.TemplateData{
				ID:  captcha.New(),
				Err: true,
			},
		})
	}

	// build absolute URLs
	//
	// Make sure you have "proxy_set_header Host $host;" besides proxy_pass in your nginx configuration

	var proto = "https"
	if strings.HasPrefix(r.Host, "127.0.") || strings.HasPrefix(r.Host, "[::1]") || strings.HasSuffix(r.Host, ".onion") { // if running locally or through TOR
		proto = "http"
	}
	var redirectURL = fmt.Sprintf("%s://%s%s", proto, r.Host, coll.Link()) // the payserver knows the collection ID anyway, and this is more convenient for the store staff than "/current"

	// refuse to create an invoice for 0 cents or so

	if coll.Due() < MaxDiscountCents {
		return errors.New("due is too low")
	}

	// create invoice

	inv, err := srv.BtcPayStore.CreateInvoice(&btcpay.InvoiceRequest{
		Amount:   float64(coll.Due()) / 100.0,
		Currency: "EUR",
		InvoiceMetadata: btcpay.InvoiceMetadata{
			OrderID: coll.ID,
		},
		InvoiceCheckout: btcpay.InvoiceCheckout{
			DefaultLanguage:   "de-DE",
			ExpirationMinutes: 60,
			MonitoringMinutes: 1440,
			RedirectURL:       redirectURL, // might be onion or clearweb
		},
	})
	if err != nil {
		return fmt.Errorf("creating invoice: %w", err)
	}

	// add event

	if err := srv.DB.CreateEvent(ordersystem.Client, coll, 0, fmt.Sprintf("Rechnung für Kryptowährungen erzeugt: [%s](%s)", inv.ID, inv.CheckoutLink)); err != nil {
		return err
	}

	// redirect to invoice

	http.Redirect(w, r, inv.CheckoutLink, http.StatusSeeOther)
	return nil
}

func (srv *Server) clientCollMessageGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("message") {
		return ErrNotFound
	}
	return html.ClientCollMessage.Execute(w, struct {
		html.TemplateData
		*ordersystem.Collection
	}{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Collection: coll,
	})
}

func (srv *Server) clientCollMessagePost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("message") {
		return ErrNotFound
	}
	if err := srv.DB.CreateEvent(ordersystem.Client, coll, 0, r.PostFormValue("message")); err != nil {
		return err
	}

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) clientCollSubmitGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("submit") {
		return ErrNotFound
	}
	return html.ClientCollSubmit.Execute(w, collView{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Actor:      ordersystem.Client,
		Collection: coll,
		ReadOnly:   true,
	})
}

func (srv *Server) clientCollSubmitPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.ClientCan("submit") {
		return ErrNotFound
	}
	if err := srv.DB.UpdateCollState(ordersystem.Client, coll, ordersystem.Submitted, 0, r.PostFormValue("submit-message")); err != nil {
		return err
	}
	srv.notify(r.Context(), "Du hast den Auftrag %s eingereicht.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

// for both Get and Post
type clientState struct {
	html.TemplateData
	Captcha   captcha.TemplateData
	CollID    string
	CollIDErr bool
	State     ordersystem.CollState
}

func (data *clientState) Valid() bool {
	if !isID(data.CollID) {
		data.CollIDErr = true
	}
	return !data.Captcha.Err && !data.CollIDErr
}

func (srv *Server) rpc(w http.ResponseWriter, r *http.Request) {

	// do verbose logging with webhook stuff
	log.Println("rpc")

	var event, err = srv.BtcPayStore.ProcessWebhook(r)
	if err != nil {
		log.Printf("error processing webhook: %v", err)
		return
	}

	log.Printf("  event: %s", event.Type)
	log.Printf("  invoice: %s", event.InvoiceID)

	// get invoice via bitpay-API, so we know the collection ID, the invoice amount and the rate at the time of payment creation

	invoice, err := srv.BitpayClient.GetInvoice(event.InvoiceID)
	if err != nil {
		log.Printf("error getting invoice: %v", err)
		return
	}

	log.Printf("  collection: %s", invoice.OrderID)

	// read collection

	coll, err := srv.DB.ReadColl(invoice.OrderID)
	if err != nil {
		log.Printf("error reading collection %s: %v", invoice.OrderID, err)
		return
	}

	// If the ordersystem books a payment in Euro (adding a log event with "paid > 0"), then we must be absolutely sure that the BtcTransmuter will sell the same amount of cryptocurrency.
	//
	// We agree that late and partial payments are sold at the exchange (by btctransmuter Fiat Conversion), but not booked in Euro.
	// Instead, the store staff must book the real selling value manually.

	// Remember that webhooks can be late or redelivered, so we must not rely on the time.Now().
	//
	// We must know which payment has been received in time because we want to book only these payments later.
	// (Imagine: pay the amount, webhook fails, crypto price drops, more coins are paid to the address, webhook gets re-delivered, shop books both payments)
	// There are two ways to know that:
	//
	// 1. AfterExpiration (only in the "InvoiceReceivedPayment" webhook) tells whether "this payment has been sent [probably more precise: received] after expiration of the invoice". Drawback: relies on successful webhook delivery (e.g. availability of btcpayserver and ordersystem)
	// 2. bitpay.Invoice.CryptoInfo.Payments.ReceivedDate, but it does not contain a time zone, probably it's UTC
	//
	// We do both. If the payment is not found in Collection.ReceivedInTimePayments or Collection.ReceivedInTimePayments, then we compare dates.

	switch event.Type {
	case btcpay.EventInvoiceReceivedPayment:
		// The "ExpirationMinutes" limit refers to this.
		//
		// Let's notify the user that her payment has been seen.
		if err := srv.invoiceReceivedPayment(event, coll, invoice); err != nil {
			log.Printf("  %v", err)
		}
	case btcpay.EventInvoiceSettled:
		// https://github.com/btcpayserver/btcpayserver/issues/2294#issuecomment-780574177
		// "once a payment is confirmed, it is considered settled"
		//
		// Conditions for settlement are:
		// - payment has been received in full
		// - payment has been received in time
		// - payment has been confirmed by the network
		//
		// In this and only this case, the BtcTransmuter must sell a similar amount of coins.
		//
		// In the "invoice settled" event, we want to book the money that has been really received, not the demanded amount.
		// We must calculate it from Invoice.CryptoInfo.
		// Risk: the hook arrives late or is redelivered manually, and late payments are added to the booking sum.
		//
		// We assume that "invoice settled" happens after "payment received" hooks.
		if err := srv.invoiceSettled(coll, invoice); err != nil {
			log.Printf("  %v", err)
		}
	default:
		log.Printf("  skipping event: %s", event.Type)
		return
	}
}

func (srv *Server) invoiceReceivedPayment(event *btcpay.InvoiceEvent, coll *ordersystem.Collection, invoice *bitpay.Invoice) error {

	// calculate fiat amount

	var paidCentsInTime int
	type cryptoAmount struct {
		Amount   float64
		Currency string
	}
	var paidLate = []cryptoAmount{} // not Euro cents, don't rely on exchange rate any more if paid late

	for _, crypto := range invoice.CryptoInfo {
		for _, payment := range crypto.Payments {
			if coll.PaymentHasBeenReceived(payment.ID) {
				continue // already in event log
			}
			if event.AfterExpiration {
				paidLate = append(paidLate, cryptoAmount{payment.Value, crypto.CryptoCode})
				coll.ReceivedLatePayments = append(coll.ReceivedLatePayments, payment.ID)
			} else {
				paidCentsInTime += int(math.Round(payment.Value * crypto.Rate * 100.0))
				coll.ReceivedInTimePayments = append(coll.ReceivedLatePayments, payment.ID)
			}
		}
	}

	// first and foremost, write modified ReceivedInTimePayments and ReceivedLatePayments to database
	if err := srv.DB.UpdateCollAndTasks(coll); err != nil {
		return fmt.Errorf("error updating collection: %v", err)
	}

	if paidCentsInTime > 0 {
		if err := srv.DB.CreateEvent(ordersystem.Bot, coll, 0, fmt.Sprintf("Rechnung [%s](%s): Vorläufiger Zahlungseingang: %s. Die Zahlung wird verbucht, sobald das Netzwerk die Transaktion bestätigt.", invoice.ID, srv.BitpayClient.InvoiceURL(invoice), html.FmtHuman(paidCentsInTime))); err != nil {
			return fmt.Errorf("error updating collection log: %v", err)
		}
	}

	for _, pl := range paidLate {
		// TODO notify store
		if err := srv.DB.CreateEvent(ordersystem.Bot, coll, 0, fmt.Sprintf("Rechnung [%s](%s): Verspäterer vorläufiger Zahlungseingang: %f %s. Da wir den Umrechnungskurs nicht mehr garantieren können, werden wir die Transaktion manuell prüfen.", invoice.ID, srv.BitpayClient.InvoiceURL(invoice), pl.Amount, pl.Currency)); err != nil {
			return fmt.Errorf("error updating collection state: %v", err)
		}
	}

	return nil
}

func (srv *Server) invoiceSettled(coll *ordersystem.Collection, invoice *bitpay.Invoice) error {

	if coll.InvoiceHasBeenBooked(invoice.ID) {
		return fmt.Errorf("invoice %s has already been booked", invoice.ID)
	}

	// calculate fiat amount

	var paidCentsInTime int

	for _, crypto := range invoice.CryptoInfo {
		for _, payment := range crypto.Payments {

			// case a: payment has been received in time and the webhook worked
			if coll.PaymentHasBeenReceivedInTime(payment.ID) {
				paidCentsInTime += int(math.Round(payment.Value * crypto.Rate * 100.0))
				continue // next payment
			}

			// case b: payment has been received late and the webhook worked
			if coll.PaymentHasBeenReceivedLate(payment.ID) {
				continue // next payment
			}

			// case c: the webhook has been missed
			recvDate, err := payment.ParseReceivedDate()
			if err != nil {
				log.Println(err)
				continue // next payment
			}
			if recvDate.Before(invoice.Expiration()) {
				paidCentsInTime += int(math.Round(payment.Value * crypto.Rate * 100.0))
			}
		}
	}

	coll.BookedInvoices = append(coll.BookedInvoices, invoice.ID)

	// first and foremost, write modified BookedInvoices to database
	if err := srv.DB.UpdateCollAndTasks(coll); err != nil {
		return fmt.Errorf("error updating collection: %v", err)
	}

	var newState ordersystem.CollState
	if paidCentsInTime+MaxDiscountCents >= coll.Due() {
		newState = ordersystem.Paid
	} else {
		newState = ordersystem.Underpaid
	}

	return srv.DB.UpdateCollState(ordersystem.Bot, coll, newState, paidCentsInTime, fmt.Sprintf("Rechnung [%s](%s): Zahlungseingang wurde bestätigt: %s.", invoice.ID, srv.BitpayClient.InvoiceURL(invoice), html.FmtHuman(paidCentsInTime)))
}

// no Collection instances involved
func (srv *Server) clientStateGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientStateGet.Execute(w, clientState{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		Captcha: captcha.TemplateData{
			ID: captcha.New(),
		},
	})
}

// no Collection instances involved
func (srv *Server) clientStatePost(w http.ResponseWriter, r *http.Request) error {

	var data = &clientState{
		TemplateData: html.TemplateData{
			TemplateData:     ssg.MakeTemplateData(srv.Langs, r),
			AuthorizedCollID: srv.sessionCollID(r),
		},
		CollID: strings.TrimSpace(r.PostFormValue("collection-id")),
	}

	if data.CollID == data.AuthorizedCollID {
		http.Redirect(w, r, fmt.Sprintf("/collection/%s", data.CollID), http.StatusSeeOther)
		return nil
	}

	if !captcha.Verify(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-answer")) {
		data.Captcha.Err = true
	}

	if !data.Valid() {
		data.Captcha.ID = captcha.New() // old captcha has been deleted during verification, so let's create a new one
		return html.ClientStateGet.Execute(w, data)
	}

	var err error
	data.State, err = srv.DB.ReadState(data.CollID)
	if err != nil {
		return err
	}

	return html.ClientStatePost.Execute(w, data)
}

// HTTP POST (not GET) for CSRF protection
func (srv *Server) clientLogoutPost(w http.ResponseWriter, r *http.Request) error {
	srv.logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (srv *Server) storeIndexGet(w http.ResponseWriter, r *http.Request) error {
	return html.StoreIndex.Execute(w, struct {
		*ordersystem.DB
		Notifications []string
	}{
		DB:            srv.DB,
		Notifications: srv.notifications(r.Context()),
	})
}

func (srv *Server) storeCollViewGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	return html.StoreCollView.Execute(w, collView{
		Actor:         ordersystem.Store,
		Collection:    coll,
		ReadOnly:      true,
		Notifications: srv.notifications(r.Context()),
	})
}

func (srv *Server) storeCollAcceptGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("accept") {
		return ErrNotFound
	}
	return html.StoreCollAccept.Execute(w, coll)
}

func (srv *Server) storeCollAcceptPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("accept") {
		return ErrNotFound
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.Accepted, 0, r.PostFormValue("accept-message")); err != nil {
		return err
	}
	srv.notify(r.Context(), "Der Auftrag %s wurde akzeptiert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollDeleteGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("delete") {
		return ErrNotFound
	}
	return html.StoreCollDelete.Execute(w, &collDelete{Collection: coll})
}

func (srv *Server) storeCollDeletePost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("delete") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-delete") == "" {
		return html.StoreCollDelete.Execute(w, &collDelete{
			Collection: coll,
			Err:        true,
		})
	}
	if err := srv.DB.Delete(ordersystem.Store, coll); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

type taskView struct {
	html.TemplateData
	*ordersystem.Task
	CollLink string
}

func (srv *Server) storeTaskConfirmArrivedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-arrived", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmArrived.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func (srv *Server) storeTaskConfirmArrivedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-arrived", task) {
		return ErrNotFound
	}
	if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Ready); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeTaskConfirmOrderedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-ordered", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmOrdered.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func (srv *Server) storeTaskConfirmOrderedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-ordered", task) {
		return ErrNotFound
	}
	if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Ordered); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeTaskMarkFailedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("mark-failed", task) {
		return ErrNotFound
	}
	return html.StoreTaskMarkFailed.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func (srv *Server) storeTaskMarkFailedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("mark-failed", task) {
		return ErrNotFound
	}
	if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Failed); err != nil {
		return err
	}
	if err := srv.DB.CreateEvent(ordersystem.Store, coll, 0, r.PostFormValue("mark-failed-message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollConfirmPayment struct {
	html.TemplateData
	Coll *ordersystem.Collection
	Err  bool
}

func (srv *Server) storeCollConfirmPaymentGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("confirm-payment") {
		return ErrNotFound
	}
	return html.StoreCollConfirmPayment.Execute(w, storeCollConfirmPayment{Coll: coll})
}

func (srv *Server) storeCollConfirmPaymentPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {

	if !coll.StoreCan("confirm-payment") {
		return ErrNotFound
	}

	var paidAmountFloat, err = strconv.ParseFloat(r.PostFormValue("paid-amount"), 64)
	if err != nil {
		return html.StoreCollConfirmPayment.Execute(w, storeCollConfirmPayment{
			Coll: coll,
			Err:  true,
		})
	}

	var paidAmount = int(math.Round(paidAmountFloat * 100.0)) // negative values are okay

	// set to underpaid or paid

	var newState ordersystem.CollState
	switch {
	case paidAmount == -1*coll.Paid():
		newState = ordersystem.Accepted
	case paidAmount >= coll.Due():
		newState = ordersystem.Paid
	default:
		newState = ordersystem.Underpaid
	}

	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, newState, paidAmount, r.PostFormValue("confirm-payment-message")); err != nil {
		return err
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollConfirmPickupGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("confirm-pickup") {
		return ErrNotFound
	}
	return html.StoreCollConfirmPickup.Execute(w, coll)
}

func (srv *Server) storeCollConfirmPickupPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {

	if !coll.StoreCan("confirm-pickup") {
		return ErrNotFound
	}

	r.ParseForm()

	for _, taskID := range r.PostForm["task[]"] {
		task, ok := coll.GetTask(taskID)
		if !ok {
			continue
		}
		if !coll.StoreCanTask("confirm-pickup", task) {
			continue
		}
		if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Fetched); err == nil {
			srv.notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)
		} else {
			return err
		}
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollConfirmReshippedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("confirm-reshipped") {
		return ErrNotFound
	}
	return html.StoreCollConfirmReshipped.Execute(w, coll)
}

func (srv *Server) storeCollConfirmReshippedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {

	if !coll.StoreCan("confirm-reshipped") {
		return ErrNotFound
	}

	r.ParseForm()

	for _, taskID := range r.PostForm["task[]"] {
		task, ok := coll.GetTask(taskID)
		if !ok {
			continue
		}
		if !coll.StoreCanTask("confirm-reshipped", task) {
			continue
		}
		if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Reshipped); err == nil {
			srv.notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)
		} else {
			return err
		}
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeTaskConfirmPickupGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-pickup", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmPickup.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func (srv *Server) storeTaskConfirmPickupPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-pickup", task) {
		return ErrNotFound
	}
	if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Fetched); err != nil {
		return err
	}
	srv.notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeTaskConfirmReshippedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-reshipped", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmReshipped.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func (srv *Server) storeTaskConfirmReshippedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection, task *ordersystem.Task) error {
	if !coll.StoreCanTask("confirm-reshipped", task) {
		return ErrNotFound
	}
	if err := srv.DB.UpdateTaskState(ordersystem.Store, task, ordersystem.Reshipped); err != nil {
		return err
	}
	srv.notify(r.Context(), "Einzelbestellung %s wurde als weiterverschickt markiert", task.ID)

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollEditGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("edit") {
		return ErrNotFound
	}
	return html.StoreCollEdit.Execute(w, collView{
		Actor:      ordersystem.Store,
		Collection: coll,
	})
}

func (srv *Server) storeCollEditPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("edit") {
		return ErrNotFound
	}
	if err := coll.MergeJSON(ordersystem.Store, r.PostFormValue("data")); err != nil {
		return err
	}

	// if the sum dropped, underpaid collection becomes paid
	if coll.State == ordersystem.Underpaid && coll.Due() <= 0 {
		coll.State = ordersystem.Paid
	}

	if err := srv.DB.UpdateCollAndTasks(coll); err != nil {
		return err
	}

	srv.notify(r.Context(), "Deine Änderungen am Auftrag %s wurden gespeichert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollMessageGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("message") {
		return ErrNotFound
	}
	return html.StoreCollMessage.Execute(w, coll)
}

func (srv *Server) storeCollMessagePost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("message") {
		return ErrNotFound
	}
	if err := srv.DB.CreateEvent(ordersystem.Store, coll, 0, r.PostFormValue("message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}
func (srv *Server) storeCollPriceRisedGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("price-rised") {
		return ErrNotFound
	}
	return html.StoreCollPriceRised.Execute(w, coll)
}

func (srv *Server) storeCollPriceRisedPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("price-rised") {
		return ErrNotFound
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.Underpaid, 0, r.PostFormValue("price-rised-message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeCollReturnGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("return") {
		return ErrNotFound
	}
	return html.StoreCollReturn.Execute(w, coll)
}

func (srv *Server) storeCollReturnPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("return") {
		return ErrNotFound
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.NeedsRevise, 0, r.PostFormValue("return-message")); err != nil {
		return err
	}
	srv.notify(r.Context(), "Der Auftrag %s wurde zur Bearbeitung zurückgegeben.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollReject struct {
	html.TemplateData
	Coll *ordersystem.Collection
	Err  bool
}

func (srv *Server) storeCollRejectGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("reject") {
		return ErrNotFound
	}
	return html.StoreCollReject.Execute(w, storeCollReject{Coll: coll})
}

func (srv *Server) storeCollRejectPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("reject") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-reject") == "" {
		return html.StoreCollReject.Execute(w, storeCollReject{
			Coll: coll,
			Err:  true,
		})
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.Rejected, 0, r.PostFormValue("reject-message")); err != nil {
		return err
	}
	srv.notify(r.Context(), "Der Auftrag %s wurde abgelehnt.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollSubmit struct {
	html.TemplateData
	Coll *ordersystem.Collection
}

func (srv *Server) storeCollSubmitGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("submit") {
		return ErrNotFound
	}
	return html.StoreCollSubmit.Execute(w, storeCollSubmit{Coll: coll})
}

func (srv *Server) storeCollSubmitPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("submit") {
		return ErrNotFound
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.Submitted, 0, r.PostFormValue("submit-message")); err != nil {
		return err
	}
	srv.notify(r.Context(), "Der Auftrag %s wurde eingereicht.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollMarkSpam struct {
	html.TemplateData
	Coll *ordersystem.Collection
	Err  bool
}

func (srv *Server) storeCollMarkSpamGet(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("mark-spam") {
		return ErrNotFound
	}
	return html.StoreCollMarkSpam.Execute(w, storeCollMarkSpam{Coll: coll})
}

func (srv *Server) storeCollMarkSpamPost(w http.ResponseWriter, r *http.Request, coll *ordersystem.Collection) error {
	if !coll.StoreCan("mark-spam") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-mark-spam") == "" {
		return html.StoreCollMarkSpam.Execute(w, storeCollMarkSpam{
			Coll: coll,
			Err:  true,
		})
	}
	if err := srv.DB.UpdateCollState(ordersystem.Store, coll, ordersystem.Spam, 0, "Dein Antrag wurde als Spam markiert."); err != nil {
		return err
	}
	srv.notify(r.Context(), "Der Auftrag %s wurde als Spam markiert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func (srv *Server) storeLoginGet(w http.ResponseWriter, r *http.Request) error {
	return html.StoreLogin.Execute(w, nil)
}

func (srv *Server) storeLoginPost(w http.ResponseWriter, r *http.Request) error {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if err := srv.Users.Authenticate(username, password); err != nil {
		return err
	}
	srv.loginStore(r.Context(), username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

type collWithPayDate struct {
	*ordersystem.Collection
	payDate string
}

func (srv *Server) storeExport(w http.ResponseWriter, r *http.Request) error {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	out := csv.NewWriter(w)
	out.Write([]string{"vat_date", "id", "country", "gross", "vat_rate", "name"})

	var colls []collWithPayDate
	for _, state := range []ordersystem.CollState{ordersystem.Accepted, ordersystem.Archived, ordersystem.Finalized, ordersystem.NeedsRevise, ordersystem.Paid, ordersystem.Submitted, ordersystem.Underpaid} {
		collIDs, err := srv.DB.ReadColls(state)
		if err != nil {
			return err
		}
		for _, collID := range collIDs {
			coll, err := srv.DB.ReadColl(collID)
			if err != nil {
				return err
			}

			var paydate string
			for _, event := range coll.Log {
				if event.NewState == ordersystem.Paid {
					paydate = string(event.Date)
					break
				}
			}
			if paydate == "" {
				continue // next collection
			}

			colls = append(colls, collWithPayDate{
				Collection: coll,
				payDate:    paydate,
			})
		}
	}

	sort.Slice(colls, func(i, j int) bool {
		if colls[i].payDate != colls[j].payDate {
			return colls[i].payDate < colls[j].payDate
		}
		return colls[i].ID < colls[j].ID
	})

	for _, coll := range colls {
		for _, task := range coll.Tasks {
			out.Flush() // before writing to w directly
			w.Write([]byte("# " + task.Merchant + "\n"))
			for _, article := range task.Articles {
				var name = article.Link
				if article.Properties != "" {
					name = name + " (" + article.Properties + ")"
				}
				out.Write([]string{coll.payDate, coll.ID, "DE", strconv.Itoa(article.Quantity * article.Price), "standard", fmt.Sprintf("%d x %s", article.Quantity, name)})
			}
		}
	}

	out.Flush()
	return nil
}

func (srv *Server) storeLogoutPost(w http.ResponseWriter, r *http.Request) error {
	srv.logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
