package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	*http.Server
}

func ListenAndServe(network, address string, handler http.Handler, stop chan os.Signal) HTTPServer {

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Println(err)
		stop <- os.Interrupt
	}

	srv := &http.Server{
		Handler:      handler,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed { // ErrServerClosed is ok
			log.Println(err)
			stop <- os.Interrupt
		}
	}()

	return HTTPServer{srv}
}

func (srv HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Server.Shutdown(ctx)
}
