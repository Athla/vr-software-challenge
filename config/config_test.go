package config_test

import (
	"testing"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestValidateConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}
