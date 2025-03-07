package server

import (
	"net/http"

	"github.com/Athla/vr-software-challenge/internal/api/handlers"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/treasury"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/Athla/vr-software-challenge/internal/service"
	"github.com/charmbracelet/log"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	limiter := tollbooth.NewLimiter(100, nil)
	r.Use(tollbooth_gin.LimitHandler(limiter))

	producer, err := messagery.NewProducer(s.cfg.Kafka.Brokers, s.cfg.Kafka.Topic)
	if err != nil {
		log.Errorf("Unable generate producer due: %s", err)
	}

	transactionHandler := handlers.TransactionHandler{
		Repo:     repository.NewTransactionRepository(s.db),
		Producer: producer,
	}

	currencyService := service.NewCurrencyService(
		treasury.NewClient(),
		repository.NewTransactionRepository(s.db),
	)

	currencyHandler := handlers.CurrencyHandler{
		CurrencyService: currencyService,
	}
	v1 := r.Group("/api/v1")
	{
		v1.GET("/transactions/:id/convert", currencyHandler.ConvertCurrency)
		v1.POST("/transactions", transactionHandler.Create)
		v1.GET("/transactions/:id", transactionHandler.GetByID)
		v1.PATCH("/transactions/:id/status", transactionHandler.UpdateStatus)
		v1.GET("/transactions", transactionHandler.List)
	}

	r.GET("/health", func(ctx *gin.Context) {
		if err := s.db.PingContext(ctx.Request.Context()); err != nil {
			log.Errorf("Unable to ping database due: %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Database connection failed."})
			return
		}

		if err := messagery.NewHealthCheck(producer, nil).Check(ctx.Request.Context()); err != nil {
			log.Errorf("Unable to check health of kafka due: %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Kafka connection failed."})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
