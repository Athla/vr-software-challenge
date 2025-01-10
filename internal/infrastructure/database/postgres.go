package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/charmbracelet/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	connString := cfg.ConnString()
	db, err := sql.Open("pgx", connString)
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

	return db, nil
}

func Transaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
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
