package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/joelazar/solve-this/internal/api"
	"github.com/joelazar/solve-this/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	server := &http.Server{
		Addr:         *addr,
		Handler:      api.NewRouter(store.New()),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on %s", *addr)
	log.Fatal(server.ListenAndServe())
}
