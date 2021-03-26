package main

import (
	"fmt"
	"log"
	"time"
)

// collection ids to be processed by the bot
var botCollIDs = make(chan string, 100)

func botColl(id string) {
	coll, err := db.ReadColl(id)
	if err != nil {
		log.Printf("bot: error reading %s: %v", id, err)
	}
	if err := db.Bot(coll); err != nil {
		log.Printf("bot: error: %s: %v", coll.ID, err)
	}
}

func bot() {
	fmt.Println("running bot")
	// get pre-transition states where actor is Bot, so we don't have to try each state
	for _, from := range db.CollFSM.From(Bot) {
		collItems, err := db.ReadColls(CollState(from))
		if err != nil {
			log.Printf("bot: error reading %s collections: %v", from, err)
			continue
		}
		for _, collItem := range collItems {
			botColl(collItem.CollID)
		}
	}
}

// Bot runs some automatic transitions.
// In order to avoid loops, it must be triggered by the ui only.
func (db *DB) Bot(coll *Collection) error {
	// if actions can be done subsequently, it's useful to put them in that order
	if err := db.BotArchive(coll); err != nil {
		return err
	}
	if err := db.BotDelete(coll); err != nil {
		return err
	}
	if err := db.BotFinalize(coll); err != nil {
		return err
	}
	// TODO process unfetched tasks
	return nil
}

func (db *DB) BotArchive(coll *Collection) error {
	if !coll.BotCan("archive") {
		return nil
	}
	since, err := coll.LatestEventSince()
	if err != nil {
		return err
	}
	if since > 2*7*24*time.Hour { // more than two weeks
		coll.ClientInput = ClientInput{} // clear data
		if err := db.UpdateCollAndTasks(coll); err != nil {
			return err
		}
		log.Printf("archiving %s", coll.ID)
		return db.UpdateCollState(Bot, coll, Archived, 0, "")
	}
	return nil
}

func (db *DB) BotDelete(coll *Collection) error {
	if !coll.BotCan("delete") {
		return nil
	}
	since, err := coll.LatestEventSince()
	if err != nil {
		return err
	}
	if since > 2*7*24*time.Hour { // more than two weeks
		log.Printf("deleting %s (%s)", coll.ID, coll.State)
		return db.Delete(Bot, coll)
	}
	return nil
}

func (db *DB) BotFinalize(coll *Collection) error {
	if !coll.BotCan("finalize") {
		return nil
	}
	if coll.Due() != 0 {
		return nil
	}
	for _, task := range coll.Tasks {
		if task.State != Fetched && task.State != Reshipped {
			// any task is neither fetched nor reshipped
			return nil
		}
	}
	log.Printf("finalizing %s", coll.ID)
	return db.UpdateCollState(Bot, coll, Finalized, 0, "Bestellauftrag ist abgeschlossen")
}
