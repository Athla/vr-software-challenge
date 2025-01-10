package server

import (
	"net/http"

	"github.com/Athla/vr-software-challenge/internal/api/handlers"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/kafka"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/charmbracelet/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	producer, err := kafka.NewProducer(s.cfg.Kafka.Brokers, s.cfg.Kafka.TransactionTopic)
	if err != nil {
		log.Errorf("Unable generate producer due: %s", err)
	}

	transactionHandler := handlers.TransactionHandler{
		Repo:     repository.NewTransactionRepository(s.db),
		Producer: producer,
	}

	r.Group("/api/v1")
	{
		r.POST("/transactions", transactionHandler.Create)
		r.GET("/transactions/:id", transactionHandler.GetByID)
		r.PATCH("/transactions/:id/status", transactionHandler.UpdateStatus)
		r.GET("/transactions", transactionHandler.List)
	}

	return r
}
