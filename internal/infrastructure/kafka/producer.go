package kafka

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producerer interface {
	PublishTransaction(ctx context.Context, msg *TransactionMessage) error
	Close()
}
type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(brokers, ","),
		"acks":               "all",
		"retries":            5,
		"retry.backoff.ms":   500,
		"linger.ms":          10,
		"compression.type":   "snappy",
		"enable.idempotence": true,
	})
	if err != nil {
		log.Errorf("Unable to create Kafka producer due: %v", err)
		return nil, err
	}
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {

					log.Errorf("Delivery in the queue failed due: %v", ev.TopicPartition.Error)
				}
			}
		}
	}()

	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *Producer) PublishTransaction(ctx context.Context, msg *TransactionMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("Unable to marshal message due: %v", err)
		return err
	}

	if err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            []byte(msg.ID.String()),
		Value:          payload,
		Headers:        []kafka.Header{{Key: "version", Value: []byte("1")}},
	}, nil); err != nil {
		log.Errorf("Unable to produce message due: %s", err)
		return err
	}

	return nil
}

func (p *Producer) Close() {
	p.producer.Flush(15 * 1000)
	p.producer.Close()
}
