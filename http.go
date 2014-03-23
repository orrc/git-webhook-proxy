package main

import (
	"log"
	"net/http"
)

func serveHttp(address string, handler http.Handler) {
	log.Println("Listening for HTTP requests at", address)
	err := http.ListenAndServe(address, handler)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTls(address, certFile, keyFile string, handler http.Handler) {
	log.Println("Listening for TLS requests at ", address)
	err := http.ListenAndServeTLS(address, certFile, keyFile, handler)
	if err != nil {
		log.Fatal(err)
	}
}
