package main

import (
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dchest/captcha"
	"github.com/dys2p/bitpay"
	"github.com/dys2p/btcpay"
	"github.com/dys2p/digitalgoods/userdb"
	"github.com/dys2p/ordersystem/html"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sethvargo/go-diceware/diceware"
)

//go:embed static
var static embed.FS

var ErrNotFound = errors.New("not found")

const MaxDiscountCents = 10

var bitpayClient *bitpay.Client
var btcpayStore btcpay.Store
var db *DB
var users userdb.Authenticator

func main() {

	log.SetFlags(0)

	// os flags

	var test = flag.Bool("test", false, "use btcpay dummy store")
	flag.Parse()

	// SQL db

	var sqlDB, err = sql.Open("sqlite3", "data/ordersystem.sqlite3?_busy_timeout=10000&_journal=WAL&_sync=NORMAL&cache=shared")
	if err != nil {
		log.Printf("error opening database: %v", err)
		return
	}

	// userdb

	users, err = userdb.Open("data/users.json")
	if err != nil {
		log.Printf("error opening userdb: %v", err)
		return
	}

	// bitpay

	bitpayClient, err = bitpay.LoadClient("data/bitpay.json")
	if err != nil {
		log.Printf("error creating bitpay API client: %v", err)
		if !*test {
			return
		}
	}

	log.Printf("please make sure that your BTCPay store is paired to the public key (hex SIN): %s", bitpayClient.SINHex())

	// btcpay

	if *test {
		btcpayStore = btcpay.NewDummyStore()
		log.Println("\033[33m" + "warning: using btcpay dummy store" + "\033[0m")
	} else {
		btcpayStore, err = btcpay.Load("data/btcpay.json")
		if err != nil {
			log.Printf("error loading btcpay store: %v", err)
			return
		}

		log.Println("don't forget to set up the webhook for your store: /rpc")
		log.Println(`  Event: "A new payment has been received"`)
		log.Println(`  Event: "An invoice has been settled"`)
	}

	// session db

	if err := initSessionManager(); err != nil {
		log.Printf("error initializing session manager: %v", err)
		return
	}

	// db

	db, err = NewDB(
		sqlDB,
		&FSM{
			Transition{State(Accepted), Bot, "confirm-payment", State(Paid)},
			Transition{State(Accepted), Bot, "confirm-payment", State(Underpaid)},
			Transition{State(Accepted), Bot, "delete", State(Deleted)},
			Transition{State(Accepted), Client, "cancel", State(Cancelled)},
			Transition{State(Accepted), Client, "pay", State(Accepted)},             // becomes Paid if payment arrives
			Transition{State(Accepted), Store, "confirm-payment", State(Paid)},      // client pays enough
			Transition{State(Accepted), Store, "confirm-payment", State(Underpaid)}, // client pays, but not enough
			Transition{State(Accepted), Store, "delete", State(Deleted)},
			Transition{State(Accepted), Store, "edit", State(Accepted)},
			Transition{State(Accepted), Store, "return", State(NeedsRevise)},
			Transition{State(Draft), Bot, "delete", State(Deleted)},
			Transition{State(Draft), Client, "delete", State(Deleted)},
			Transition{State(Draft), Client, "edit", State(Draft)},
			Transition{State(Draft), Client, "submit", State(Submitted)},
			Transition{State(Finalized), Store, "message", State(Finalized)}, // "Hi, we just shipped your order."
			Transition{State(Finalized), Bot, "archive", State(Archived)},
			Transition{State(NeedsRevise), Client, "cancel", State(Cancelled)},
			Transition{State(NeedsRevise), Client, "edit", State(NeedsRevise)},
			Transition{State(NeedsRevise), Client, "submit", State(Submitted)},
			Transition{State(Paid), Bot, "confirm-payment", State(Paid)},
			Transition{State(Paid), Bot, "finalize", State(Finalized)},
			Transition{State(Paid), Store, "confirm-payment", State(Paid)}, // refund overpaid amount
			Transition{State(Paid), Store, "confirm-pickup", State(Paid)},
			Transition{State(Paid), Store, "confirm-reshipped", State(Paid)},
			Transition{State(Paid), Store, "edit", State(Paid)}, // price or availability changed after payment
			Transition{State(Paid), Store, "message", State(Paid)},
			Transition{State(Paid), Store, "price-rised", State(Underpaid)},
			Transition{State(Spam), Bot, "delete", State(Deleted)},
			Transition{State(Submitted), Client, "cancel", State(Cancelled)},
			Transition{State(Submitted), Store, "accept", State(Accepted)},
			Transition{State(Submitted), Store, "edit", State(Submitted)},
			Transition{State(Submitted), Store, "mark-spam", State(Spam)},
			Transition{State(Submitted), Store, "reject", State(Rejected)},
			Transition{State(Submitted), Store, "return", State(NeedsRevise)},
			Transition{State(Underpaid), Bot, "confirm-payment", State(Paid)},
			Transition{State(Underpaid), Bot, "confirm-payment", State(Underpaid)},
			Transition{State(Underpaid), Client, "message", State(Underpaid)},
			Transition{State(Underpaid), Client, "pay", State(Underpaid)},            // becomes Paid if payment arrives
			Transition{State(Underpaid), Store, "confirm-payment", State(Paid)},      // client pays missing amount
			Transition{State(Underpaid), Store, "confirm-payment", State(Underpaid)}, // client pays a part of the missing amount
			Transition{State(Underpaid), Store, "edit", State(Paid)},                 // store modifies the collection, the sum drops, paid sum is now enough
			Transition{State(Underpaid), Store, "edit", State(Underpaid)},            // store modifies the collection, but it is still underpaid
			Transition{State(Underpaid), Store, "message", State(Underpaid)},
		},
		&FSM{
			Transition{State(NotOrderedYet), Store, "confirm-ordered", State(Ordered)},
			Transition{State(NotOrderedYet), Store, "mark-failed", State(Failed)},
			Transition{State(Ordered), Store, "confirm-arrived", State(Ready)},
			Transition{State(Ordered), Store, "mark-failed", State(Failed)},
			Transition{State(Ready), Bot, "pickup-expired", State(Unfetched)}, // TODO
			Transition{State(Ready), Store, "confirm-pickup", State(Fetched)},
			Transition{State(Ready), Store, "confirm-reshipped", State(Reshipped)},
		},
	)
	if err != nil {
		log.Printf("error creating database: %v", err)
		return
	}

	// bot

	var ticker = time.NewTicker(12 * time.Hour)
	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-ticker.C:
				wg.Add(1)
				bot()
				wg.Done()
			case id := <-botCollIDs:
				wg.Add(1)
				botColl(id)
				wg.Done()
			}
		}
	}()

	bot() // run now

	// http handlers

	var stop = make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var substatic, _ = fs.Sub(fs.FS(static), "static")

	var clientRouter = httprouter.New()
	clientRouter.ServeFiles("/static/*filepath", http.FS(substatic))
	clientRouter.HandlerFunc(http.MethodGet, "/", client(clientHelloGet))
	clientRouter.HandlerFunc(http.MethodGet, "/create", client(clientCreateGet))
	clientRouter.HandlerFunc(http.MethodPost, "/create", client(clientCreatePost))
	clientRouter.HandlerFunc(http.MethodGet, "/current", client(clientCollCurrentGet))
	clientRouter.HandlerFunc(http.MethodGet, "/collection", client(clientCollLoginGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection", client(clientCollLoginPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid", clientWithCollection(clientCollViewGet))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/cancel", clientWithCollection(clientCollCancelGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/cancel", clientWithCollection(clientCollCancelPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/delete", clientWithCollection(clientCollDeleteGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/delete", clientWithCollection(clientCollDeletePost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/edit", clientWithCollection(clientCollEditGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/edit", clientWithCollection(clientCollEditPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/message", clientWithCollection(clientCollMessageGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/message", clientWithCollection(clientCollMessagePost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/pay-btcpay", clientWithCollection(clientCollPayBTCPayGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/pay-btcpay", clientWithCollection(clientCollPayBTCPayPost))
	clientRouter.HandlerFunc(http.MethodGet, "/collection/:collid/submit", clientWithCollection(clientCollSubmitGet))
	clientRouter.HandlerFunc(http.MethodPost, "/collection/:collid/submit", clientWithCollection(clientCollSubmitPost))
	clientRouter.HandlerFunc(http.MethodPost, "/rpc", rpc)
	clientRouter.HandlerFunc(http.MethodGet, "/state", client(clientStateGet))
	clientRouter.HandlerFunc(http.MethodPost, "/state", client(clientStatePost))
	clientRouter.HandlerFunc(http.MethodPost, "/logout", client(clientLogoutPost))
	clientRouter.Handler("GET", "/captcha/:fn", captcha.Server(captcha.StdWidth, captcha.StdHeight))

	var clientSrv = ListenAndServe("tcp", ":9000", sessionManager.LoadAndSave(clientRouter), stop)

	var storeRouter = httprouter.New()
	storeRouter.ServeFiles("/static/*filepath", http.FS(substatic))
	storeRouter.HandlerFunc(http.MethodGet, "/login", store(storeLoginGet))
	storeRouter.HandlerFunc(http.MethodPost, "/login", store(storeLoginPost))
	// with authentication:
	storeRouter.HandlerFunc(http.MethodGet, "/", auth(store(storeIndexGet)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid", auth(storeWithCollection(storeCollViewGet)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/accept", auth(storeWithCollection(storeCollAcceptGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/accept", auth(storeWithCollection(storeCollAcceptPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-payment", auth(storeWithCollection(storeCollConfirmPaymentGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-payment", auth(storeWithCollection(storeCollConfirmPaymentPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-pickup", auth(storeWithCollection(storeCollConfirmPickupGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-pickup", auth(storeWithCollection(storeCollConfirmPickupPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-pickup/:taskid", auth(storeWithTask(storeTaskConfirmPickupGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-pickup/:taskid", auth(storeWithTask(storeTaskConfirmPickupPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-reshipped", auth(storeWithCollection(storeCollConfirmReshippedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-reshipped", auth(storeWithCollection(storeCollConfirmReshippedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-reshipped/:taskid", auth(storeWithTask(storeTaskConfirmReshippedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-reshipped/:taskid", auth(storeWithTask(storeTaskConfirmReshippedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/delete", auth(storeWithCollection(storeCollDeleteGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/delete", auth(storeWithCollection(storeCollDeletePost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/edit", auth(storeWithCollection(storeCollEditGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/edit", auth(storeWithCollection(storeCollEditPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/mark-spam", auth(storeWithCollection(storeCollMarkSpamGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/mark-spam", auth(storeWithCollection(storeCollMarkSpamPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/message", auth(storeWithCollection(storeCollMessageGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/message", auth(storeWithCollection(storeCollMessagePost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/price-rised", auth(storeWithCollection(storeCollPriceRisedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/price-rised", auth(storeWithCollection(storeCollPriceRisedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/return", auth(storeWithCollection(storeCollReturnGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/return", auth(storeWithCollection(storeCollReturnPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/reject", auth(storeWithCollection(storeCollRejectGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/reject", auth(storeWithCollection(storeCollRejectPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-arrived/:taskid", auth(storeWithTask(storeTaskConfirmArrivedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-arrived/:taskid", auth(storeWithTask(storeTaskConfirmArrivedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/confirm-ordered/:taskid", auth(storeWithTask(storeTaskConfirmOrderedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/confirm-ordered/:taskid", auth(storeWithTask(storeTaskConfirmOrderedPost)))
	storeRouter.HandlerFunc(http.MethodGet, "/collection/:collid/mark-failed/:taskid", auth(storeWithTask(storeTaskMarkFailedGet)))
	storeRouter.HandlerFunc(http.MethodPost, "/collection/:collid/mark-failed/:taskid", auth(storeWithTask(storeTaskMarkFailedPost)))
	storeRouter.HandlerFunc(http.MethodPost, "/logout", store(storeLogoutPost))

	var storeSrv = ListenAndServe("tcp", "127.0.0.1:9001", sessionManager.LoadAndSave(storeRouter), stop)

	// run until we receive an interrupt or any listener fails

	log.Printf("running")
	<-stop // blocks
	log.Println("shutting down")
	clientSrv.Shutdown()
	storeSrv.Shutdown()
	wg.Wait()
}

type collView struct {
	*Collection
	Actor         Actor
	ReadOnly      bool
	ShowHints     bool
	Notifications []string
}

type clientHello struct {
	AuthorizedCollID string
	Notifications    []string
}

func clientHelloGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientHello.Execute(w, clientHello{
		AuthorizedCollID: sessionCollID(r),
		Notifications:    notifications(r.Context()),
	})
}

type clientCreate struct {
	CollID              string
	CollIDErr           bool
	CollPass            string
	CollPassErr         bool
	CaptchaID           string
	CaptchaErr          bool
	CheckWrittenDown    bool
	CheckWrittenDownErr bool

	AuthorizedCollID string
}

func (data *clientCreate) Valid() bool {
	if !IsID(data.CollID) {
		data.CollIDErr = true
	}
	if data.CollPass == "" {
		data.CollPassErr = true
	}
	if !data.CheckWrittenDown {
		data.CheckWrittenDownErr = true
	}
	return !data.CollIDErr && !data.CollPassErr && !data.CaptchaErr && !data.CheckWrittenDownErr
}

func clientCreateGet(w http.ResponseWriter, r *http.Request) error {
	var collPass, err = diceware.GenerateWithWordList(5, WordListGermanSmall)
	if err != nil {
		return err
	}
	return html.ClientCreate.Execute(w, &clientCreate{
		CollID:           NewID(),
		CollPass:         strings.Join(collPass, "-"),
		CaptchaID:        captcha.NewLen(6),
		AuthorizedCollID: sessionCollID(r),
	})
}

func clientCreatePost(w http.ResponseWriter, r *http.Request) error {

	var data = &clientCreate{
		CollID:           strings.TrimSpace(r.PostFormValue("collection-id")),
		CollPass:         strings.TrimSpace(r.PostFormValue("collection-passphrase")),
		CheckWrittenDown: r.PostFormValue("check-written-down") != "",
		AuthorizedCollID: sessionCollID(r),
	}

	if !captcha.VerifyString(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-solution")) {
		data.CaptchaErr = true
	}

	if !data.Valid() {
		data.CaptchaID = captcha.New() // old captcha has been deleted during verification, so let's create a new one
		return html.ClientCreate.Execute(w, data)
	}

	var coll = &Collection{
		ID:    data.CollID,
		State: Draft,
	}

	if hash, err := HashPassword(data.CollPass); err == nil {
		coll.Pass = string(hash)
	} else {
		return err
	}

	if err := db.CreateCollection(coll); err != nil {
		return err
	}

	loginClient(r.Context(), coll.ID)
	http.Redirect(w, r, fmt.Sprintf("/collection/%s/edit", coll.ID), http.StatusSeeOther)
	return nil
}

func clientCollEditGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("edit") {
		return ErrNotFound
	}
	return html.ClientCollEdit.Execute(w, collView{
		Actor:      Client,
		Collection: coll,
		ShowHints:  true,
	})
}

func clientCollEditPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("edit") {
		return ErrNotFound
	}
	if err := coll.MergeJSON(Client, r.PostFormValue("data")); err != nil {
		return err
	}
	if err := db.UpdateCollAndTasks(coll); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type clientLogin struct {
	CollID      string
	CollIDErr   bool
	CollPassErr bool

	AuthorizedCollID string
}

// success and error url for payment providers, so we don't reveal the collection ID to them
func clientCollCurrentGet(w http.ResponseWriter, r *http.Request) error {
	if collID := sessionCollID(r); collID != "" {
		http.Redirect(w, r, fmt.Sprintf("/collection/%s", collID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	return nil
}

func clientCollLoginGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientCollLogin.Execute(w, &clientLogin{
		AuthorizedCollID: sessionCollID(r),
	})
}

func clientCollLoginPost(w http.ResponseWriter, r *http.Request) error {
	var id = strings.TrimSpace(r.FormValue("collection-id"))
	var data = &clientLogin{
		CollID:           id,
		AuthorizedCollID: sessionCollID(r),
	}
	if !IsID(id) {
		data.CollIDErr = true
		return html.ClientCollLogin.Execute(w, data)
	}
	var pass = strings.TrimSpace(r.PostFormValue("collection-passphrase"))
	var coll, err = db.ReadCollPass(id, pass)
	if err != nil {
		data.CollPassErr = true
		return html.ClientCollLogin.Execute(w, data)
	}

	loginClient(r.Context(), coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func clientCollViewGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	return html.ClientCollView.Execute(w, collView{
		Actor:         Client,
		Collection:    coll,
		ReadOnly:      true,
		Notifications: notifications(r.Context()),
	})
}

type clientCollCancel struct {
	*Collection
	Err bool
}

func clientCollCancelGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("cancel") {
		return ErrNotFound
	}
	return html.ClientCollCancel.Execute(w, &clientCollCancel{Collection: coll})
}

func clientCollCancelPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("cancel") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-cancel") == "" {
		return html.ClientCollCancel.Execute(w, &clientCollCancel{
			Collection: coll,
			Err:        true,
		})
	}
	if err := db.UpdateCollState(Client, coll, Cancelled, 0, ""); err != nil {
		return err
	}

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

// used by client and store
type collDelete struct {
	*Collection
	Err bool
}

func clientCollDeleteGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("delete") {
		return ErrNotFound
	}
	return html.ClientCollDelete.Execute(w, &collDelete{
		Collection: coll,
	})
}

func clientCollDeletePost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("delete") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-delete") == "" {
		return html.ClientCollDelete.Execute(w, &collDelete{
			Collection: coll,
			Err:        true,
		})
	}
	if err := db.Delete(Client, coll); err != nil {
		return err
	}

	logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

type clientCollPayBTCPay struct {
	*Collection
	CaptchaID  string
	CaptchaErr bool
}

func clientCollPayBTCPayGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("pay") {
		return ErrNotFound
	}
	return html.ClientCollPayBTCPay.Execute(w, &clientCollPayBTCPay{
		Collection: coll,
		CaptchaID:  captcha.New(),
	})
}

func clientCollPayBTCPayPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {

	if !coll.ClientCan("pay") {
		return ErrNotFound
	}
	if !captcha.VerifyString(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-solution")) {
		return html.ClientCollPayBTCPay.Execute(w, &clientCollPayBTCPay{
			Collection: coll,
			CaptchaErr: true,
			CaptchaID:  captcha.New(),
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

	inv, err := btcpayStore.CreateInvoice(&btcpay.InvoiceRequest{
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

	if err := db.CreateEvent(Client, coll, 0, fmt.Sprintf("Rechnung für Kryptowährungen erzeugt: [%s](%s)", inv.ID, inv.CheckoutLink)); err != nil {
		return err
	}

	// redirect to invoice

	http.Redirect(w, r, inv.CheckoutLink, http.StatusSeeOther)
	return nil
}

func clientCollMessageGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("message") {
		return ErrNotFound
	}
	return html.ClientCollMessage.Execute(w, coll)
}

func clientCollMessagePost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("message") {
		return ErrNotFound
	}
	if err := db.CreateEvent(Client, coll, 0, r.PostFormValue("message")); err != nil {
		return err
	}

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func clientCollSubmitGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("submit") {
		return ErrNotFound
	}
	return html.ClientCollSubmit.Execute(w, collView{
		Actor:      Client,
		Collection: coll,
		ReadOnly:   true,
	})
}

func clientCollSubmitPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.ClientCan("submit") {
		return ErrNotFound
	}
	if err := db.UpdateCollState(Client, coll, Submitted, 0, r.PostFormValue("submit-message")); err != nil {
		return err
	}
	notify(r.Context(), "Du hast den Auftrag %s eingereicht.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

// for both Get and Post
type clientState struct {
	CaptchaID  string
	CaptchaErr bool
	CollID     string
	CollIDErr  bool
	State      CollState

	AuthorizedCollID string
}

func (data *clientState) Valid() bool {
	if !IsID(data.CollID) {
		data.CollIDErr = true
	}
	return !data.CaptchaErr && !data.CollIDErr
}

func rpc(w http.ResponseWriter, r *http.Request) {

	// do verbose logging with webhook stuff
	log.Println("rpc")

	var event, err = btcpayStore.ProcessWebhook(r)
	if err != nil {
		log.Printf("error processing webhook: %v", err)
		return
	}

	log.Printf("  event: %s", event.Type)
	log.Printf("  invoice: %s", event.InvoiceID)

	// get invoice via bitpay-API, so we know the collection ID, the invoice amount and the rate at the time of payment creation

	invoice, err := bitpayClient.GetInvoice(event.InvoiceID)
	if err != nil {
		log.Printf("error getting invoice: %v", err)
		return
	}

	log.Printf("  collection: %s", invoice.OrderID)

	// read collection

	coll, err := db.ReadColl(invoice.OrderID)
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
		if err := invoiceReceivedPayment(event, coll, invoice); err != nil {
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
		if err := invoiceSettled(coll, invoice); err != nil {
			log.Printf("  %v", err)
		}
	default:
		log.Printf("  skipping event: %s", event.Type)
		return
	}
}

func invoiceReceivedPayment(event *btcpay.InvoiceEvent, coll *Collection, invoice *bitpay.Invoice) error {

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
	if err := db.UpdateCollAndTasks(coll); err != nil {
		return fmt.Errorf("error updating collection: %v", err)
	}

	if paidCentsInTime > 0 {
		if err := db.CreateEvent(Bot, coll, 0, fmt.Sprintf("Rechnung [%s](%s): Vorläufiger Zahlungseingang: %s. Die Zahlung wird verbucht, sobald das Netzwerk die Transaktion bestätigt.", invoice.ID, bitpayClient.InvoiceURL(invoice), html.FmtHuman(paidCentsInTime))); err != nil {
			return fmt.Errorf("error updating collection log: %v", err)
		}
	}

	for _, pl := range paidLate {
		// TODO notify store
		if err := db.CreateEvent(Bot, coll, 0, fmt.Sprintf("Rechnung [%s](%s): Verspäterer vorläufiger Zahlungseingang: %f %s. Da wir den Umrechnungskurs nicht mehr garantieren können, werden wir die Transaktion manuell prüfen.", invoice.ID, bitpayClient.InvoiceURL(invoice), pl.Amount, pl.Currency)); err != nil {
			return fmt.Errorf("error updating collection state: %v", err)
		}
	}

	return nil
}

func invoiceSettled(coll *Collection, invoice *bitpay.Invoice) error {

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
	if err := db.UpdateCollAndTasks(coll); err != nil {
		return fmt.Errorf("error updating collection: %v", err)
	}

	var newState CollState
	if paidCentsInTime+MaxDiscountCents >= coll.Due() {
		newState = Paid
	} else {
		newState = Underpaid
	}

	return db.UpdateCollState(Bot, coll, newState, paidCentsInTime, fmt.Sprintf("Rechnung [%s](%s): Zahlungseingang wurde bestätigt: %s.", invoice.ID, bitpayClient.InvoiceURL(invoice), html.FmtHuman(paidCentsInTime)))
}

// no Collection instances involved
func clientStateGet(w http.ResponseWriter, r *http.Request) error {
	return html.ClientStateGet.Execute(w, clientState{
		AuthorizedCollID: sessionCollID(r),
		CaptchaID:        captcha.New(),
	})
}

// no Collection instances involved
func clientStatePost(w http.ResponseWriter, r *http.Request) error {

	var data = &clientState{
		AuthorizedCollID: sessionCollID(r),
		CollID:           strings.TrimSpace(r.PostFormValue("collection-id")),
	}

	if data.CollID == data.AuthorizedCollID {
		http.Redirect(w, r, fmt.Sprintf("/collection/%s", data.CollID), http.StatusSeeOther)
		return nil
	}

	if !captcha.VerifyString(r.PostFormValue("captcha-id"), r.PostFormValue("captcha-solution")) {
		data.CaptchaErr = true
	}

	if !data.Valid() {
		data.CaptchaID = captcha.New() // old captcha has been deleted during verification, so let's create a new one
		return html.ClientStateGet.Execute(w, data)
	}

	var err error
	data.State, err = db.ReadState(data.CollID)
	if err != nil {
		return err
	}

	return html.ClientStatePost.Execute(w, data)
}

// HTTP POST (not GET) for CSRF protection
func clientLogoutPost(w http.ResponseWriter, r *http.Request) error {
	logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func storeIndexGet(w http.ResponseWriter, r *http.Request) error {
	return html.StoreIndex.Execute(w, struct {
		*DB
		Notifications []string
	}{
		db,
		notifications(r.Context()),
	})
}

func storeCollViewGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	return html.StoreCollView.Execute(w, collView{
		Actor:         Store,
		Collection:    coll,
		ReadOnly:      true,
		Notifications: notifications(r.Context()),
	})
}

func storeCollAcceptGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("accept") {
		return ErrNotFound
	}
	return html.StoreCollAccept.Execute(w, coll)
}

func storeCollAcceptPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("accept") {
		return ErrNotFound
	}
	if err := db.UpdateCollState(Store, coll, Accepted, 0, r.PostFormValue("accept-message")); err != nil {
		return err
	}
	notify(r.Context(), "Der Auftrag %s wurde akzeptiert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollDeleteGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("delete") {
		return ErrNotFound
	}
	return html.StoreCollDelete.Execute(w, &collDelete{Collection: coll})
}

func storeCollDeletePost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("delete") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-delete") == "" {
		return html.StoreCollDelete.Execute(w, &collDelete{
			Collection: coll,
			Err:        true,
		})
	}
	if err := db.Delete(Store, coll); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

type taskView struct {
	*Task
	CollLink string
}

func storeTaskConfirmArrivedGet(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-arrived", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmArrived.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func storeTaskConfirmArrivedPost(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-arrived", task) {
		return ErrNotFound
	}
	if err := db.UpdateTaskState(Store, task, Ready); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeTaskConfirmOrderedGet(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-ordered", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmOrdered.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func storeTaskConfirmOrderedPost(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-ordered", task) {
		return ErrNotFound
	}
	if err := db.UpdateTaskState(Store, task, Ordered); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeTaskMarkFailedGet(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("mark-failed", task) {
		return ErrNotFound
	}
	return html.StoreTaskMarkFailed.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func storeTaskMarkFailedPost(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("mark-failed", task) {
		return ErrNotFound
	}
	if err := db.UpdateTaskState(Store, task, Failed); err != nil {
		return err
	}
	if err := db.CreateEvent(Store, coll, 0, r.PostFormValue("mark-failed-message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollConfirmPayment struct {
	Coll *Collection
	Err  bool
}

func storeCollConfirmPaymentGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("confirm-payment") {
		return ErrNotFound
	}
	return html.StoreCollConfirmPayment.Execute(w, storeCollConfirmPayment{Coll: coll})
}

func storeCollConfirmPaymentPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {

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

	var newState CollState
	if paidAmount >= coll.Due() {
		newState = Paid
	} else {
		newState = Underpaid
	}

	if err := db.UpdateCollState(Store, coll, newState, paidAmount, r.PostFormValue("confirm-payment-message")); err != nil {
		return err
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollConfirmPickupGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("confirm-pickup") {
		return ErrNotFound
	}
	return html.StoreCollConfirmPickup.Execute(w, coll)
}

func storeCollConfirmPickupPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {

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
		if err := db.UpdateTaskState(Store, task, Fetched); err == nil {
			notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)
		} else {
			return err
		}
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollConfirmReshippedGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("confirm-reshipped") {
		return ErrNotFound
	}
	return html.StoreCollConfirmReshipped.Execute(w, coll)
}

func storeCollConfirmReshippedPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {

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
		if err := db.UpdateTaskState(Store, task, Reshipped); err == nil {
			notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)
		} else {
			return err
		}
	}

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeTaskConfirmPickupGet(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-pickup", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmPickup.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func storeTaskConfirmPickupPost(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-pickup", task) {
		return ErrNotFound
	}
	if err := db.UpdateTaskState(Store, task, Fetched); err != nil {
		return err
	}
	notify(r.Context(), "Einzelbestellung %s wurde als abgeholt markiert", task.ID)

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeTaskConfirmReshippedGet(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-reshipped", task) {
		return ErrNotFound
	}
	return html.StoreTaskConfirmReshipped.Execute(w, taskView{
		Task:     task,
		CollLink: coll.Link(),
	})
}

func storeTaskConfirmReshippedPost(w http.ResponseWriter, r *http.Request, coll *Collection, task *Task) error {
	if !coll.StoreCanTask("confirm-reshipped", task) {
		return ErrNotFound
	}
	if err := db.UpdateTaskState(Store, task, Reshipped); err != nil {
		return err
	}
	notify(r.Context(), "Einzelbestellung %s wurde als weiterverschickt markiert", task.ID)

	botCollIDs <- coll.ID

	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollEditGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("edit") {
		return ErrNotFound
	}
	return html.StoreCollEdit.Execute(w, collView{
		Actor:      Store,
		Collection: coll,
	})
}

func storeCollEditPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("edit") {
		return ErrNotFound
	}
	if err := coll.MergeJSON(Store, r.PostFormValue("data")); err != nil {
		return err
	}

	// if the sum dropped, underpaid collection becomes paid
	if coll.State == Underpaid && coll.Due() <= 0 {
		coll.State = Paid
	}

	if err := db.UpdateCollAndTasks(coll); err != nil {
		return err
	}

	notify(r.Context(), "Deine Änderungen am Auftrag %s wurden gespeichert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollMessageGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("message") {
		return ErrNotFound
	}
	return html.StoreCollMessage.Execute(w, coll)
}

func storeCollMessagePost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("message") {
		return ErrNotFound
	}
	if err := db.CreateEvent(Store, coll, 0, r.PostFormValue("message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}
func storeCollPriceRisedGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("price-rised") {
		return ErrNotFound
	}
	return html.StoreCollPriceRised.Execute(w, coll)
}

func storeCollPriceRisedPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("price-rised") {
		return ErrNotFound
	}
	if err := db.UpdateCollState(Store, coll, Underpaid, 0, r.PostFormValue("price-rised-message")); err != nil {
		return err
	}
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeCollReturnGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("return") {
		return ErrNotFound
	}
	return html.StoreCollReturn.Execute(w, coll)
}

func storeCollReturnPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("return") {
		return ErrNotFound
	}
	if err := db.UpdateCollState(Store, coll, NeedsRevise, 0, r.PostFormValue("return-message")); err != nil {
		return err
	}
	notify(r.Context(), "Der Auftrag %s wurde zur Bearbeitung zurückgegeben.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollReject struct {
	Coll *Collection
	Err  bool
}

func storeCollRejectGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("reject") {
		return ErrNotFound
	}
	return html.StoreCollReject.Execute(w, storeCollReject{Coll: coll})
}

func storeCollRejectPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("reject") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-reject") == "" {
		return html.StoreCollReject.Execute(w, storeCollReject{
			Coll: coll,
			Err:  true,
		})
	}
	if err := db.UpdateCollState(Store, coll, Rejected, 0, r.PostFormValue("reject-message")); err != nil {
		return err
	}
	notify(r.Context(), "Der Auftrag %s wurde abgelehnt.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

type storeCollMarkSpam struct {
	Coll *Collection
	Err  bool
}

func storeCollMarkSpamGet(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("mark-spam") {
		return ErrNotFound
	}
	return html.StoreCollMarkSpam.Execute(w, storeCollMarkSpam{Coll: coll})
}

func storeCollMarkSpamPost(w http.ResponseWriter, r *http.Request, coll *Collection) error {
	if !coll.StoreCan("mark-spam") {
		return ErrNotFound
	}
	if r.PostFormValue("confirm-mark-spam") == "" {
		return html.StoreCollMarkSpam.Execute(w, storeCollMarkSpam{
			Coll: coll,
			Err:  true,
		})
	}
	if err := db.UpdateCollState(Store, coll, Spam, 0, "Dein Antrag wurde als Spam markiert."); err != nil {
		return err
	}
	notify(r.Context(), "Der Auftrag %s wurde als Spam markiert.", coll.ID)
	http.Redirect(w, r, coll.Link(), http.StatusSeeOther)
	return nil
}

func storeLoginGet(w http.ResponseWriter, r *http.Request) error {
	return html.StoreLogin.Execute(w, nil)
}

func storeLoginPost(w http.ResponseWriter, r *http.Request) error {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if err := users.Authenticate(username, password); err != nil {
		return err
	}
	loginStore(r.Context(), username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func storeLogoutPost(w http.ResponseWriter, r *http.Request) error {
	logout(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
