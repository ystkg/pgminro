package main

import (
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

// config
const (
	host        = "localhost"
	defaultPort = "8432"

	defaultPgx = false

	useHttpsCertainty = false // ex. by reverse proxy

	// avoid unlimited
	sessionTimeoutMin = 180
	maxIdleTimeMin    = 10
	queryTimeoutSec   = 300
	maxRows           = 10000
)

func main() {
	port := defaultPort
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(host+":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
