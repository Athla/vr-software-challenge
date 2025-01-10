package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/api/server"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/database"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/Athla/vr-software-challenge/migrations"
	"github.com/charmbracelet/log"
	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Info("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Info("Server exiting")

	done <- true
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := migrations.Migrate(db, context.Background()); err != nil {
		log.Fatalf("Unable to run migrations: %v", err)
	}

	// Initialize repositories
	txRepo := repository.NewTransactionRepository(db)

	// Initialize Kafka producer
	producer, err := messagery.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Fatalf("Unable to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize consumer with handler
	consumerHandler := func(ctx context.Context, msg *messagery.TransactionMessage) error {
		return txRepo.UpdateStatus(ctx, msg.ID, models.StatusCompleted)
	}

	consumer, err := messagery.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
		cfg.Kafka.Topic,
		consumerHandler,
	)
	if err != nil {
		log.Fatalf("Unable to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Start consumer in background
	go func() {
		if err := consumer.Start(context.Background()); err != nil {
			log.Errorf("Consumer error: %v", err)
		}
	}()

	// Initialize HTTP server
	server := server.NewServer(cfg, db, producer)

	done := make(chan bool, 1)
	go gracefulShutdown(server, done)

	log.Infof("Server starting on :%d", cfg.App.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	<-done
	log.Info("Server stopped gracefully")
}
