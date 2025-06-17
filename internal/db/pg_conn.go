package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgresConn() (*Postgres, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("PG_HOST"),
	os.Getenv("PG_PORT"),
	os.Getenv("PG_USER"),
	os.Getenv("PG_PASSWORD"),		
	os.Getenv("PG_DBNAME"),
)

	db, err := sql.Open("postgres", dsn	)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(15 * time.Minute)
	db.SetConnMaxLifetime(1 * time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil{
		log.Fatal(err)
	}
	 
	return &Postgres	{db: db}, nil	
}