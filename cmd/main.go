package main

import (
	"log"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/db"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/repository"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	"github.com/Guizzs26/personal-blog/internal/server"
	"github.com/Guizzs26/personal-blog/pkg/cronx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger.SetupLogger()

	conn := db.NewPostgresConn()
	srv := server.NewServer(conn)

	userRepo := repository.NewPostgresUserRepository(conn)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(conn)
	authService := service.NewAuthService(userRepo, refreshTokenRepo)

	if err := cronx.StartCleanupCronJob(authService); err != nil {
		log.Fatalf("Failed to start cleanup cron job: %v", err)
	}

	log.Println("Starting server on :4444")
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
