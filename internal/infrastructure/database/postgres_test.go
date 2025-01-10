package database_test

import (
	"testing"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
)

func TestNewConnection(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "checkout_user",
		Password: "checkout_password",
		Database: "vr_checkout_db",
		SSLMode:  "disable",
	}
	dbHandler, err := database.NewConnection(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, dbHandler)
}
