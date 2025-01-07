package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/charmbracelet/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type DbHandler struct {
	db  *sql.DB
	ctx context.Context
}

func NewConnection(cfg config.DatabaseConfig) (*DbHandler, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Errorf("Unable to open database due: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		log.Errorf("Database is not alive due: %v", err)
		return nil, err
	}

	return &DbHandler{
		db: db,
	}, nil
}

func (db *DbHandler) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.db.BeginTx(db.ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		log.Errorf("Unable to create transaction due: %v", err)
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Errorf("Unable to perform transaction due: '%v'.\n\tOriginal error: '%v'", rbErr, err)
			return rbErr
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("Unable to commit transaction due: %v", err)
		return err
	}

	return nil
}
