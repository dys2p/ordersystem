package ordersystem

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("not found")

type DB struct {
	sqlDB *sql.DB

	// collection
	createColl      *sql.Stmt
	readColl        *sql.Stmt
	readColls       *sql.Stmt
	readState       *sql.Stmt
	updateColl      *sql.Stmt
	updateCollState *sql.Stmt
	deleteColl      *sql.Stmt

	// event
	createEvent  *sql.Stmt
	readEvents   *sql.Stmt
	deleteEvents *sql.Stmt

	// task
	createTask      *sql.Stmt
	readTasks       *sql.Stmt
	updateTaskState *sql.Stmt
	deleteTasks     *sql.Stmt
}

func NewDB(sqlDB *sql.DB) (*DB, error) {

	var db = &DB{
		sqlDB: sqlDB,
	}

	_, err := sqlDB.Exec(`
		create table if not exists coll (
			id    text primary key,
			pass  text not null,
			state text not null,
			data  text not null,

			client_contact           text not null,
			client_contact_protocol  text not null,
			delivery_first_name      text not null,
			delivery_last_name       text not null,
			delivery_addr_supplement text not null,
			delivery_customer_id     text not null, -- e. g. DHL PostNumber
			delivery_street          text not null,
			delivery_housenumber     text not null,
			delivery_postcode        text not null,
			delivery_city            text not null,
			delivery_email           text not null,
			delivery_phone           text not null,
			delivery_tracking_ids    text not null,

			country                  text not null,
			delivery_method          text not null,
			delivery_gross_price     int  not null,
			shipping_service         text not null
		);
		create table if not exists event (
			id        integer primary key,
			collid    text not null,
			collstate text not null,
			date      text not null,
			paid      INTEGER NOT NULL,
			text      text not null
		);
		create table if not exists task (
			id     text primary key,
			collid text not null,
			state  text not null,
			data   text not null
		);
	`)
	if err != nil {
		return nil, err
	}

	// collection

	db.createColl, err = db.sqlDB.Prepare("insert into coll (id, pass, state, data, client_contact, client_contact_protocol, delivery_first_name, delivery_last_name, delivery_addr_supplement, delivery_customer_id, delivery_street, delivery_housenumber, delivery_postcode, delivery_city, delivery_email, delivery_phone, delivery_tracking_ids, country, delivery_method, delivery_gross_price, shipping_service) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	db.readColl, err = db.sqlDB.Prepare("select pass, state, data, client_contact, client_contact_protocol, delivery_first_name, delivery_last_name, delivery_addr_supplement, delivery_customer_id, delivery_street, delivery_housenumber, delivery_postcode, delivery_city, delivery_email, delivery_phone, delivery_tracking_ids, country, delivery_method, delivery_gross_price, shipping_service from coll where id = ? limit 1")
	if err != nil {
		return nil, err
	}

	db.readColls, err = db.sqlDB.Prepare("select id from coll where coll.state = ?")
	if err != nil {
		return nil, err
	}

	db.readState, err = db.sqlDB.Prepare("select state from coll where id = ? limit 1")
	if err != nil {
		return nil, err
	}

	db.updateColl, err = db.sqlDB.Prepare("update coll set data = ?, client_contact = ?, client_contact_protocol = ?, delivery_first_name = ?, delivery_last_name = ?, delivery_addr_supplement = ?, delivery_customer_id = ?, delivery_street = ?, delivery_housenumber = ?, delivery_postcode = ?, delivery_city = ?, delivery_email = ?, delivery_phone = ?, delivery_tracking_ids = ?, country = ?, delivery_method = ?, delivery_gross_price = ?, shipping_service = ? where id = ?")
	if err != nil {
		return nil, err
	}

	db.updateCollState, err = db.sqlDB.Prepare("update coll set state = ? where id = ?")
	if err != nil {
		return nil, err
	}

	db.deleteColl, err = db.sqlDB.Prepare("delete from coll where id = ?")
	if err != nil {
		return nil, err
	}

	// event

	db.createEvent, err = db.sqlDB.Prepare("insert into event (collid, collstate, date, paid, text) values (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	db.readEvents, err = db.sqlDB.Prepare("select collstate, date, paid, text FROM event where collid = ? order by id desc")
	if err != nil {
		return nil, err
	}

	db.deleteEvents, err = db.sqlDB.Prepare("delete from event where collid = ?")
	if err != nil {
		return nil, err
	}

	// task

	db.createTask, err = db.sqlDB.Prepare("insert or replace into task (id, collid, state, data) values (?, ?, ?, ?)") // not upsert, which is useful for partial updates but not required here
	if err != nil {
		return nil, err
	}

	db.readTasks, err = db.sqlDB.Prepare("select id, state, data from task where collid = ?")
	if err != nil {
		return nil, err
	}

	db.updateTaskState, err = db.sqlDB.Prepare("update task set state = ? where id = ?")
	if err != nil {
		return nil, err
	}

	db.deleteTasks, err = db.sqlDB.Prepare("delete from task where collid = ?")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) CreateCollection(coll *Collection) error {

	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no effect after commit

	if _, err := tx.Stmt(db.createColl).Exec(coll.ID, coll.Pass, Draft, "{}", "", "", "", "", "", "", "", "", "", "", "", "", "[]", "DE", "", 0, ""); err != nil {
		return err
	}

	var firstEvent = Event{ // required for the bot which deletes old drafts
		NewState: Draft,
		Date:     Today(),
		Paid:     0,
		Text:     "Auftragsentwurf wurde angelegt",
	}
	coll.Log = []Event{firstEvent}

	if _, err := tx.Stmt(db.createEvent).Exec(coll.ID, firstEvent.NewState, firstEvent.Date, firstEvent.Paid, firstEvent.Text); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateEvent creates an event. UpdateCollState should be preferred if the collection state changes.
func (db *DB) CreateEvent(actor Actor, coll *Collection, paid int, message string) error {

	message = strings.TrimSpace(message)
	if message != "" {
		message = fmt.Sprintf("%s: %s", actor.Name(), message)
	}

	_, err := db.createEvent.Exec(coll.ID, coll.State, Today(), paid, message)
	return err
}

func (db *DB) Delete(actor Actor, coll *Collection) error {

	if !CollFSM.Can(actor, State(coll.State), State(Deleted)) {
		return errors.New("not allowed to delete collection") // deletion is important, so we must state clearly if it fails (and not just return ErrNotFound)
	}

	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no effect after commit

	if _, err := tx.Stmt(db.deleteColl).Exec(coll.ID); err != nil {
		return err
	}
	if _, err := tx.Stmt(db.deleteEvents).Exec(coll.ID); err != nil {
		return err
	}
	if _, err := tx.Stmt(db.deleteTasks).Exec(coll.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) ReadColl(id string) (*Collection, error) {

	var collData string
	var deliveryTrackingIDs string // json array
	var coll = &Collection{ID: id}
	if err := db.readColl.QueryRow(id).Scan(&coll.Pass, &coll.State, &collData, &coll.ClientContact, &coll.ClientContactProtocol, &coll.DeliveryAddress.FirstName, &coll.DeliveryAddress.LastName, &coll.DeliveryAddress.Supplement, &coll.DeliveryAddress.CustomerID, &coll.DeliveryAddress.Street, &coll.DeliveryAddress.HouseNumber, &coll.DeliveryAddress.Postcode, &coll.DeliveryAddress.City, &coll.DeliveryAddress.Email, &coll.DeliveryAddress.Phone, &deliveryTrackingIDs, &coll.CountryID, &coll.DeliveryMethodID, &coll.DeliveryGrossPrice, &coll.ShippingServiceID); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(collData), &coll.CollectionData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(deliveryTrackingIDs), &coll.DeliveryTrackingIDs); err != nil {
		return nil, fmt.Errorf("unmarshaling %s: %w", deliveryTrackingIDs, err)
	}

	// events

	events, err := db.readEvents.Query(id)
	if err != nil {
		return nil, err
	}
	defer events.Close()

	for events.Next() {
		var event = Event{}
		if err := events.Scan(&event.NewState, &event.Date, &event.Paid, &event.Text); err != nil {
			return nil, err
		}
		coll.Log = append(coll.Log, event)
	}

	// tasks

	tasks, err := db.readTasks.Query(id)
	if err != nil {
		return nil, err
	}
	defer tasks.Close()

	for tasks.Next() {
		var taskData string
		var task = &Task{}
		if err := tasks.Scan(&task.ID, &task.State, &taskData); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(taskData), &task.TaskData); err != nil {
			return nil, err
		}
		coll.Tasks = append(coll.Tasks, task)
	}

	return coll, nil
}

func (db *DB) ReadCollPass(id, pass string) (*Collection, error) {
	var coll, err = db.ReadColl(id)
	if err != nil {
		return nil, err
	}
	if !coll.CompareHash(pass) {
		return nil, ErrNotFound
	}
	return coll, nil
}

// task count is currently filtered with NotOrderedYet
func (db *DB) ReadColls(state CollState) ([]string, error) {
	rows, err := db.readColls.Query(state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (db *DB) ReadState(id string) (CollState, error) {
	var state string
	return CollState(state), db.readState.QueryRow(id).Scan(&state)
}

// coll must contain the old state
func (db *DB) UpdateCollState(actor Actor, coll *Collection, newState CollState, paidAmount int, message string) error {

	if !CollFSM.Can(actor, State(coll.State), State(newState)) {
		return ErrNotFound
	}

	message = strings.TrimSpace(message)
	if message != "" {
		message = fmt.Sprintf("%s: %s", actor.Name(), message)
	}

	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no effect after commit

	if _, err := tx.Stmt(db.updateCollState).Exec(newState, coll.ID); err != nil {
		return err
	}

	if _, err := tx.Stmt(db.createEvent).Exec(coll.ID, newState, Today(), paidAmount, message); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	coll.State = newState
	return nil
}

// updates the collection given by coll.ID
func (db *DB) UpdateCollAndTasks(coll *Collection) error {

	data, err := json.Marshal(coll.CollectionData)
	if err != nil {
		return err
	}
	deliveryTrackingIDs, err := json.Marshal(coll.DeliveryTrackingIDs)
	if err != nil {
		return err
	}

	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no effect after commit

	if _, err := tx.Stmt(db.updateColl).Exec(string(data), coll.ClientContact, coll.ClientContactProtocol, coll.DeliveryAddress.FirstName, coll.DeliveryAddress.LastName, coll.DeliveryAddress.Supplement, coll.DeliveryAddress.CustomerID, coll.DeliveryAddress.Street, coll.DeliveryAddress.HouseNumber, coll.DeliveryAddress.Postcode, coll.DeliveryAddress.City, coll.DeliveryAddress.Email, coll.DeliveryAddress.Phone, deliveryTrackingIDs, coll.CountryID, coll.DeliveryMethodID, coll.DeliveryGrossPrice, coll.ShippingServiceID, coll.ID); err != nil {
		return err
	}

	if _, err := tx.Stmt(db.deleteTasks).Exec(coll.ID); err != nil {
		return err
	}

	for _, task := range coll.Tasks {
		if task.State == "" {
			task.State = NotOrderedYet // initial state
		}
		taskData, err := json.Marshal(task.TaskData)
		if err != nil {
			return err
		}
		if _, err := tx.Stmt(db.createTask).Exec(task.ID, coll.ID, task.State, string(taskData)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// task must contain the old state
func (db *DB) UpdateTaskState(actor Actor, task *Task, newState TaskState) error {
	if !TaskFSM.Can(actor, State(task.State), State(newState)) {
		return ErrNotFound
	}
	if _, err := db.updateTaskState.Exec(newState, task.ID); err != nil {
		return err
	}
	task.State = newState
	return nil
}
