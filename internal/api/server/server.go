package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port     int
	cfg      *config.Config
	db       *sql.DB
	producer messagery.Producerer
}

func NewServer(cfg *config.Config, db *sql.DB, producer messagery.Producerer) *http.Server {
	server := &Server{
		port:     cfg.App.Port,
		cfg:      cfg,
		db:       db,
		producer: producer,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
