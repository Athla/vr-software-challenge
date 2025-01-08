package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagePublished = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_published_total",
		Help: "The total number of published messages",
	})

	messagePublishErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_publish_errors_total",
		Help: "The total number of message publish errors",
	})

	messageProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_processed_total",
		Help: "The total number of processed messages",
	})
)
