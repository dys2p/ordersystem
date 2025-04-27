package main

import (
	"fmt"
	"log"

	"github.com/dys2p/ordersystem"
)

// collection ids to be processed by the bot
var botCollIDs = make(chan string, 100)

func (srv *Server) BotColl(id string) {
	coll, err := srv.DB.ReadColl(id)
	if err != nil {
		log.Printf("bot: error reading %s: %v", id, err)
	}
	if err := srv.DB.Bot(coll); err != nil {
		log.Printf("bot: error: %s: %v", coll.ID, err)
	}
}

func (srv *Server) Bot() {
	fmt.Println("running bot")
	// get pre-transition states where actor is Bot, so we don't have to try each state
	for _, from := range ordersystem.CollFSM.From(ordersystem.Bot) {
		collIDs, err := srv.DB.ReadColls(ordersystem.CollState(from))
		if err != nil {
			log.Printf("bot: error reading %s collections: %v", from, err)
			continue
		}
		for _, collID := range collIDs {
			srv.BotColl(collID)
		}
	}
}
