package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type DB struct {
	sqlDB *sql.DB

	CollFSM *FSM
	TaskFSM *FSM

	// collection
	createColl      *sql.Stmt
	readColl        *sql.Stmt
	readColls       *sql.Stmt
	readState       *sql.Stmt
	updateCollData  *sql.Stmt
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

func NewDB(sqlDB *sql.DB, collFSM, taskFSM *FSM) (*DB, error) {

	var db = &DB{
		sqlDB:   sqlDB,
		CollFSM: collFSM,
		TaskFSM: taskFSM,
	}

	_, err := sqlDB.Exec(`
		create table if not exists coll (
			id    text primary key,
			pass  text not null,
			state text not null,
			data  text not null
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

	db.createColl, err = db.sqlDB.Prepare("insert into coll (id, pass, state, data) values (?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	db.readColl, err = db.sqlDB.Prepare("select pass, state, data from coll where id = ? limit 1")
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

	db.updateCollData, err = db.sqlDB.Prepare("update coll set data = ? where id = ?")
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

	if _, err := tx.Stmt(db.createColl).Exec(coll.ID, coll.Pass, Draft, "{}"); err != nil {
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

	if !db.CollFSM.Can(actor, State(coll.State), State(Deleted)) {
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

	var coll = &Collection{ID: id}
	if err := db.readColl.QueryRow(id).Scan(&coll.Pass, &coll.State, &collData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(collData), &coll.CollectionData); err != nil {
		return nil, err
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

	if !db.CollFSM.Can(actor, State(coll.State), State(newState)) {
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

	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no effect after commit

	if _, err := tx.Stmt(db.updateCollData).Exec(string(data), coll.ID); err != nil {
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
	if !db.TaskFSM.Can(actor, State(task.State), State(newState)) {
		return ErrNotFound
	}
	if _, err := db.updateTaskState.Exec(newState, task.ID); err != nil {
		return err
	}
	task.State = newState
	return nil
}
