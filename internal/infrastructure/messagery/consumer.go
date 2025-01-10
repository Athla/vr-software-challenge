package messagery

import (
	"context"
	"strings"
	"time"

	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MessageHandler func(context.Context, *TransactionMessage) error

type Consumer struct {
	consumer *kafka.Consumer
	handler  MessageHandler
	topic    string
}

func NewConsumer(brokers []string, groupID, topic string, handler MessageHandler) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    strings.Join(brokers, ","),
		"group.id":             groupID,
		"auto.offset.reset":    "earliest",
		"enable.auto.commit":   false,
		"max.poll.interval.ms": 300000,
		"session.timeout.ms":   45000,
	})

	if err != nil {
		log.Errorf("Unable to create consumer due: %s", err)
		return nil, err
	}

	return &Consumer{
		consumer: c,
		handler:  handler,
		topic:    topic,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if err := c.consumer.SubscribeTopics([]string{c.topic}, nil); err != nil {
		log.Errorf("Unable to subscribe to topics due: %s", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.consumer.ReadMessage(time.Millisecond)
			if err != nil {
				if err.(kafka.Error).IsFatal() {
					log.Errorf("FATAL ERROR. Unable to read from consumer due: %s", err)
					return err
				}
				log.Warnf("Unable to read message from consumer due: %s", err)
				continue

			}
			var transaction TransactionMessage
			if err := json.Unmarshal(msg.Value, &transaction); err != nil {
				log.Warnf("Unable to unmarshal error due: %s", err)
				continue
			}
			if err := c.handler(ctx, &transaction); err != nil {
				log.Warnf("Unable to handle message due: %s", err)
				continue
			}

			if _, err = c.consumer.CommitMessage(msg); err != nil {
				log.Warnf("Unable to commit message due: %s", err)
				continue
			}

		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
