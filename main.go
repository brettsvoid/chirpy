package main

import (
	"log"
	"net/http"
	"time"
)

const PORT = "8080"

func main() {
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:           ":" + PORT,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Listening on port: %s\n", PORT)
	log.Fatal(s.ListenAndServe())
}
