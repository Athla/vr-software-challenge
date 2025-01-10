package main

import (
	"context"
	"log"
	"os"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/database"
	"github.com/Athla/vr-software-challenge/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
		os.Exit(1)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Migrate(db, context.Background()); err != nil {
		log.Fatalf("Migration failed: %v", err)
		os.Exit(1)
	}

	log.Println("Migrations completed successfully")
}
