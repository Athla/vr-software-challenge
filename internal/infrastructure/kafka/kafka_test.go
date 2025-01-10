package kafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/internal/infrastructure/kafka"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPublishTransaction(t *testing.T) {
	producer, err := kafka.NewProducer([]string{"localhost:9092"}, "transactions")
	assert.NoError(t, err)
	assert.NotNil(t, producer)

	msg := &kafka.TransactionMessage{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		CreatedAt:       time.Now(),
	}

	err = producer.PublishTransaction(context.Background(), msg)
	assert.NoError(t, err)
}
