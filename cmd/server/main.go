package main

import (
	"log"

	"github.com/zaynkorai/enlabs/internal/app/server"
	"github.com/zaynkorai/enlabs/internal/app/services"
	"github.com/zaynkorai/enlabs/internal/platform/persistence"
	"github.com/zaynkorai/enlabs/internal/transport/http"
	"github.com/zaynkorai/enlabs/pkg/config"
	"github.com/zaynkorai/enlabs/pkg/database"
)

// @title Enlabs Balance Processing API
// @version 1.0
// @description API for processing incoming requests from 3rd-party providers and managing user balances.
// @host localhost:8089

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, closeErr := db.DB()
		if closeErr != nil {
			log.Printf("Error getting *sql.DB from GORM: %v", closeErr)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	userRepo := persistence.NewUserRepository(db)
	transactionRepo := persistence.NewTransactionRepository(db)

	transactionService := services.NewTransactionService(userRepo, transactionRepo)

	httpHandler := http.NewHandler(transactionService)

	srv := server.NewServer(cfg, httpHandler)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
