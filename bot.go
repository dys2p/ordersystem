package ordersystem

import (
	"log"
	"time"

	"github.com/dys2p/eco/delivery"
)

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
	if since > 2*7*24*time.Hour && coll.Due() == 0 { // more than two weeks and we're even
		coll.ClientContact = ""
		coll.ClientContactProtocol = ""
		coll.DeliveryAddress = delivery.Address{}
		coll.DeliveryTrackingIDs = nil
		// keep DeliveryMethod and country
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
	if coll.Due() > 0 {
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
