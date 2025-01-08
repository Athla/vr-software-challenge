package kafka

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type HealthCheck struct {
	producer *Producer
	consumer *Consumer
}

func NewHealthCheck(producer *Producer, consumer *Consumer) *HealthCheck {
	return &HealthCheck{
		producer: producer,
		consumer: consumer,
	}
}

func (h *HealthCheck) Check(ctx context.Context) error {
	testMsg := &TransactionMessage{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}

	if err := h.producer.PublishTransaction(ctx, testMsg); err != nil {
		log.Errorf("Unable to publish health transaction: %s", err)
		return err
	}

	return nil
}
