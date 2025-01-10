package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

// @title           VR Software Challenge API
// @version         1.0
// @description     A transaction processing API with asynchronous queue processing
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  guilher.c.rodrigues@gmail.com

// @license.name   Unlicense
// @license.url    https://unlicense.org/

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	if err := migrations.Migrate(db, ctx); err != nil {
		log.Fatalf("Unable to run migrations: %v", err)
	}

	txRepo := repository.NewTransactionRepository(db)

	producer, err := messagery.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Fatalf("Unable to create Kafka producer: %v", err)
	}

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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("Consumer error: %v", err)
		}
	}()

	server := server.NewServer(cfg, db, producer)

	serverShutdown := make(chan struct{})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
		close(serverShutdown)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	if err := consumer.Close(); err != nil {
		log.Errorf("Error closing consumer: %v", err)
	}

	producer.Close()

	wg.Wait()

	<-serverShutdown

	log.Info("Server exited properly")
}
