package messagery

import (
	"context"
	"sync"
)

type MockProducer struct {
	mu       sync.Mutex
	messages []TransactionMessage
}

func NewMockProducer() Producerer {
	return &MockProducer{
		messages: make([]TransactionMessage, 0),
	}
}

func (m *MockProducer) PublishTransaction(ctx context.Context, msg *TransactionMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, *msg)
	return nil
}

func (m *MockProducer) Close() {
	// No-op for mock
}
