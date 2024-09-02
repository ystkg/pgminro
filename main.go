package main

import (
	"log"
	"net/http"
)

// config
const (
	host        = "localhost"
	defaultPort = "8432"

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
