package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Collection struct {
	// stored as SQL row, but partly unmarshaled from user input, hence the JSON tags
	ID    string    `json:"-"`
	Pass  string    `json:"-"`
	State CollState `json:"-"`
	CollectionData
	Log   []Event  `json:"-"`
	Tasks TaskList `json:"tasks"`
}

// AuthorizedCollID returns the collection ID.
// This is useful in templates, where we can pass a Collection or an "AuthorizedCollID" struct field.
func (coll *Collection) AuthorizedCollID() string {
	return coll.ID
}

func (coll *Collection) BotCan(action string) bool {
	return db.CollFSM.CanAction(Bot, State(coll.State), action)
}

func (coll *Collection) ClientCan(action string) bool {
	return db.CollFSM.CanAction(Client, State(coll.State), action)
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (coll *Collection) CompareHash(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(coll.Pass), []byte(password)) == nil
}

func (coll *Collection) GetTask(id string) (*Task, bool) {
	// linear search, okay for a small amount of tasks
	for _, task := range coll.Tasks {
		if task.ID == id {
			return task, true
		}
	}
	return nil, false
}

func (coll *Collection) LatestEventSince() (time.Duration, error) {
	var latest time.Time
	for _, event := range coll.Log {
		t, err := event.Date.Parse()
		if err != nil {
			return 0, err
		}
		if t.After(latest) {
			latest = t
		}
	}
	return time.Since(latest), nil
}

// Link returns an absolute URL without host, like "/collection/ABCDEFGHKL"
func (coll *Collection) Link() string {
	return fmt.Sprintf("/collection/%s", coll.ID)
}

// MergeJSON unmarshals JSON data and calls Merge.
func (coll *Collection) MergeJSON(actor Actor, untrustedColl string) error {
	var uc = &Collection{}
	if err := json.Unmarshal([]byte(untrustedColl), uc); err != nil {
		return fmt.Errorf("unmarshaling user input: %w", err)
	}
	return coll.Merge(actor, uc)
}

// Merge merges an untrusted collection into the receiving collection.
// coll.ID is never modified.
func (coll *Collection) Merge(actor Actor, untrustedColl *Collection) error {

	// untrustedColl.ID might be manipulated, overwriting mitigates the event of accidentally using it
	untrustedColl.ID = coll.ID

	// copy collection data
	//
	// We must take care to not existing data with empty data.

	switch actor {
	case Client:
		coll.ClientInput = untrustedColl.ClientInput
		// don't modify coll.StoreInput
	case Store:
		coll.ClientInput = untrustedColl.ClientInput
		coll.StoreInput = untrustedColl.StoreInput
	}

	// If missing, assign IDs to (probably new) untrusted tasks.
	//
	// A malicious client could have set the task number as well. It is important that they can't overwrite read-only tasks. TaskList.Clear will check that.

	for _, task := range untrustedColl.Tasks {
		if strings.TrimSpace(task.ID) == "" {
			task.ID = NewID()
		}
	}

	// merge tasks

	coll.Tasks.Clear(actor)

	for _, newTask := range untrustedColl.Tasks {
		coll.Tasks.Insert(newTask)

		if len(coll.Tasks) > 100 {
			return errors.New("too many tasks")
		}
	}

	return nil
}

func (coll *Collection) Due() int {
	return coll.Sum() - coll.Paid()
}

func (coll *Collection) Paid() int {
	var sum = 0
	for _, event := range coll.Log {
		sum += event.Paid
	}
	return sum
}

func (coll *Collection) Sum() int {
	var sum = 0
	for _, task := range coll.Tasks {
		sum += task.Sum()
		sum += task.Fee()
	}

	// same logic as in ordersystem.js updateView()
	if coll.DeliveryMethod == "shipping" {

		// 1. default value
		var reshippingFee int
		for _, s := range coll.ShippingServices() {
			if s.ID == coll.ShippingServiceID {
				reshippingFee = s.MinCost
			}
		}

		// 2. data value
		if coll.ReshippingFee != 0 {
			reshippingFee = coll.ReshippingFee
		}

		sum += reshippingFee
	}

	return sum
}

func (coll *Collection) StoreCan(action string) bool {
	return db.CollFSM.CanAction(Store, State(coll.State), action)
}

func (coll *Collection) StoreCanTask(action string, task *Task) bool {
	if coll.State == Paid || coll.State == Underpaid {
		// if underpaid, store must assess the risk
		return db.TaskFSM.CanAction(Store, State(task.State), action)
	}
	return false
}

// CollectionData is a separate struct so we can marshal it easily and store it in the SQL database.
type CollectionData struct {
	ClientInput
	StoreInput
	BookedInvoices         []string `json:"booked-invoices"`           // bitpay.Invoice.ID, booking is triggered by the invoice settled webhook
	ReceivedInTimePayments []string `json:"received-in-time-payments"` // bitpay.Invoice.InvoiceData.CryptoInfo.Payments.ID, event log like "Vorläufiger Zahlungseingang"
	ReceivedLatePayments   []string `json:"received-late-payments"`    // bitpay.Invoice.InvoiceData.CryptoInfo.Payments.ID, event log like "Verspäterer vorläufiger Zahlungseingang"
}

func (data *CollectionData) InvoiceHasBeenBooked(invoiceID string) bool {
	for _, id := range data.BookedInvoices {
		if id == invoiceID {
			return true
		}
	}
	return false
}

func (data *CollectionData) PaymentHasBeenReceived(paymentID string) bool {
	return data.PaymentHasBeenReceivedInTime(paymentID) || data.PaymentHasBeenReceivedLate(paymentID)
}

func (data *CollectionData) PaymentHasBeenReceivedInTime(paymentID string) bool {
	for _, id := range data.ReceivedInTimePayments {
		if id == paymentID {
			return true
		}
	}
	return false
}

func (data *CollectionData) PaymentHasBeenReceivedLate(paymentID string) bool {
	for _, id := range data.ReceivedLatePayments {
		if id == paymentID {
			return true
		}
	}
	return false
}

type ClientInput struct {
	ClientName                string `json:"client-name"`
	ClientContact             string `json:"client-contact"`
	ClientContactProtocol     string `json:"client-contact-protocol"`
	DeliveryMethod            string `json:"delivery-method"`
	ShippingAddressSupplement string `json:"shipping-address-supplement"`
	ShippingFirstName         string `json:"shipping-first-name"`
	ShippingLastName          string `json:"shipping-last-name"`
	ShippingPostcode          string `json:"shipping-postcode"`
	ShippingServiceID         string `json:"shipping-service"`
	ShippingStreet            string `json:"shipping-street"`
	ShippingStreetNumber      string `json:"shipping-street-number"`
	ShippingTown              string `json:"shipping-town"`
}

type ContactProtocol struct {
	ID   string
	Name string
}

// ClientContactProtocols returns available contact protocols. If an unknown (i.e. deprecated) protocol is stored in the database, it is returned as well.
func (data *ClientInput) ContactProtocols() []ContactProtocol {
	var protocols = []ContactProtocol{
		ContactProtocol{"email", "E-Mail"},
		ContactProtocol{"xmpp-otr", "Jabber mit OTR"},
		ContactProtocol{"matrix", "Matrix"},
		ContactProtocol{"signal", "Signal"},
	}
	if data.ClientContactProtocol == "" {
		return protocols
	}
	// append data.ClientContactProtocol if it's not among them (i.e. if it's deprecated)
	var ok = false
	for _, p := range protocols {
		if p.ID == data.ClientContactProtocol {
			ok = true
		}
	}
	if !ok {
		protocols = append(protocols, ContactProtocol{data.ClientContactProtocol, data.ClientContactProtocol})
	}
	return protocols
}

type ShippingService struct {
	ID      string
	Name    string
	MinCost int
}

// ShippingServices returns available shipping services. If an unknown (i.e. deprecated) service is stored in the database, it is returned as well.
func (data *ClientInput) ShippingServices() []ShippingService {
	var services = []ShippingService{
		ShippingService{"dhl-paket-analog", "DHL Paket (versichert), analog frankiert: ab 7,49 €", 749},
		ShippingService{"dhl-paket-digital", "DHL Paket (versichert), digital frankiert: ab 5,94 €", 594},  // 499 * 1.19
		ShippingService{"post-einschreiben-einwurf", "Deutsche Post Einschreiben Einwurf: ab 3,81 €", 381}, // (85 + 235) * 1.19
		ShippingService{"post-einschreiben-wert", "Deutsche Post Einschreiben Wert: ab 6,31 €", 631},       // (85 + 445) * 1.19
	}
	if data.ShippingServiceID == "" {
		return services
	}
	// append data.ShippingServiceID if it's not among them (i.e. if it's deprecated)
	var ok = false
	for _, s := range services {
		if s.ID == data.ShippingServiceID {
			ok = true
		}
	}
	if !ok {
		services = append(services, ShippingService{data.ShippingServiceID, data.ShippingServiceID, 0})
	}
	return services
}

type StoreInput struct {
	ReshippingFee int `json:"reshipping-fee"`
}

type TaskList []*Task

func (tl *TaskList) Clear(actor Actor) {
	// https://github.com/golang/go/wiki/SliceTricks#filter-in-place
	n := 0
	for _, task := range *tl {
		if !task.Writeable(actor) {
			// task is not writeable, keep it
			(*tl)[n] = task
			n++
		}
	}
	*tl = (*tl)[:n]
}

func (tl *TaskList) Insert(task *Task) {
	// linear search, slow but okay for small numbers of tasks
	for _, t := range *tl {
		if t.ID == task.ID {
			return // not necessarily an error, the task might just be read-only
		}
	}
	*tl = append(*tl, task)
}
