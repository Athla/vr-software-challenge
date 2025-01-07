package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/database"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
	cfg  *config.Config
	db   *database.DbHandler
}

func NewServer(cfg *config.Config) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db, _ := database.NewConnection(cfg.Database)
	NewServer := &Server{
		port: port,
		cfg:  cfg,
		db:   db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
