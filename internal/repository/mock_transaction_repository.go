package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/google/uuid"
)

type MockTransactionRepository struct {
	mu           sync.Mutex
	transactions map[uuid.UUID]*models.Transaction
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[uuid.UUID]*models.Transaction),
	}
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx.CreatedAt = time.Now()
	m.transactions[tx.ID] = tx
	return nil
}

func (m *MockTransactionRepository) GetById(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, exists := m.transactions[id]
	if !exists {
		return nil, errors.ErrTransactionNotFound
	}
	return tx, nil
}

func (m *MockTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, exists := m.transactions[id]
	if !exists {
		return errors.ErrTransactionNotFound
	}

	tx.Status = status
	if status == models.StatusCompleted {
		now := time.Now()
		tx.ProcessedAt = &now
	}
	return nil
}

func (m *MockTransactionRepository) List(ctx context.Context, limit, offset int) ([]models.Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var transactions []models.Transaction
	for _, tx := range m.transactions {
		transactions = append(transactions, *tx)
	}

	// Simulate pagination
	start := offset
	end := offset + limit
	if start > len(transactions) {
		start = len(transactions)
	}
	if end > len(transactions) {
		end = len(transactions)
	}

	return transactions[start:end], nil
}
