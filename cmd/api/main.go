package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":4444",
		Handler: mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Println("Starting server on :4444")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}	