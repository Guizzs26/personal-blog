package main

import (
	"log"

	"github.com/Guizzs26/personal-blog/internal/db"
	"github.com/Guizzs26/personal-blog/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn := db.NewPostgresConn()
	srv := server.NewServer(conn)

	log.Println("Starting server on :4444")
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
