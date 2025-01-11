package service

import (
	"context"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/treasury"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyService_ConvertTransaction(t *testing.T) {
	now := time.Now()
	mockTreasury := treasury.NewMockClient()
	mockRepo := repository.NewMockTransactionRepository()

	// Create test transactions
	validTx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Valid Transaction",
		TransactionDate: now,
		AmountUSD:       decimal.NewFromFloat(100.00),
		Status:          models.StatusCompleted,
	}

	oldTx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Old Transaction",
		TransactionDate: now.AddDate(-1, 0, 0), // 1 year old
		AmountUSD:       decimal.NewFromFloat(100.00),
		Status:          models.StatusCompleted,
	}

	// Store transactions in mock repository
	mockRepo.Create(context.Background(), validTx)
	mockRepo.Create(context.Background(), oldTx)

	mockTreasury.AddMockRate("OLD_RATE", decimal.NewFromFloat(1.5), now.AddDate(-1, 0, 0))

	service := NewCurrencyService(mockTreasury, mockRepo)

	tests := []struct {
		name           string
		setupTx        func() uuid.UUID
		targetCurrency string
		expectedError  error
	}{
		{
			name: "Valid EUR conversion",
			setupTx: func() uuid.UUID {
				return validTx.ID
			},
			targetCurrency: "EUR",
			expectedError:  nil,
		},
		{
			name: "Valid GBP conversion",
			setupTx: func() uuid.UUID {
				return validTx.ID
			},
			targetCurrency: "GBP",
			expectedError:  nil,
		},
		{
			name: "Invalid currency",
			setupTx: func() uuid.UUID {
				return validTx.ID
			},
			targetCurrency: "INVALID",
			expectedError:  errors.ErrInvalidCurrency,
		},
		{
			name: "Empty currency",
			setupTx: func() uuid.UUID {
				return validTx.ID
			},
			targetCurrency: "",
			expectedError:  errors.ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txID := tt.setupTx()
			conversion, err := service.ConvertTransaction(
				context.Background(),
				txID,
				tt.targetCurrency,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, conversion)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conversion)
				assert.Equal(t, txID, conversion.TransactionID)
				assert.Equal(t, tt.targetCurrency, conversion.TargetCurrency)

				assert.False(t, conversion.ExchangeDate.After(conversion.TransactionDate))
				assert.False(t, conversion.ExchangeDate.Before(conversion.TransactionDate.AddDate(0, -6, 0)))

				if txID == validTx.ID {
					assert.Equal(t, validTx.Description, conversion.Description)
					assert.Equal(t, validTx.AmountUSD.String(), conversion.OriginalAmount.String())
				}
			}
		})
	}
}

func TestCurrencyService_ConvertTransaction_Rounding(t *testing.T) {
	now := time.Now().UTC()
	mockTreasury := treasury.NewMockClient()
	mockRepo := repository.NewMockTransactionRepository()

	rate := decimal.NewFromFloat(0.8333)
	mockTreasury.AddMockRate("EUR", rate, now.Add(-24*time.Hour))

	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Rounding Test",
		TransactionDate: now,
		AmountUSD:       decimal.NewFromFloat(100.77),
		Status:          models.StatusCompleted,
	}

	mockRepo.Create(context.Background(), tx)
	service := NewCurrencyService(mockTreasury, mockRepo)

	conversion, err := service.ConvertTransaction(
		context.Background(),
		tx.ID,
		"EUR",
	)

	assert.NoError(t, err)
	assert.NotNil(t, conversion)
}
