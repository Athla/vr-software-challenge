package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewCurrencyConversion(t *testing.T) {
	tx := &Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.00),
	}

	tests := []struct {
		name           string
		transaction    *Transaction
		targetCurrency string
		exchangeRate   decimal.Decimal
		exchangeDate   time.Time
		wantErr        bool
	}{
		{
			name:           "Valid conversion",
			transaction:    tx,
			targetCurrency: "EUR",
			exchangeRate:   decimal.NewFromFloat(0.85),
			exchangeDate:   tx.TransactionDate,
			wantErr:        false,
		},
		{
			name:           "Exchange date too old",
			transaction:    tx,
			targetCurrency: "EUR",
			exchangeRate:   decimal.NewFromFloat(0.85),
			exchangeDate:   tx.TransactionDate.AddDate(0, -7, 0),
			wantErr:        true,
		},
		{
			name:           "Future exchange date",
			transaction:    tx,
			targetCurrency: "EUR",
			exchangeRate:   decimal.NewFromFloat(0.85),
			exchangeDate:   tx.TransactionDate.AddDate(0, 0, 1),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conversion, err := NewCurrencyConversion(
				tt.transaction,
				tt.targetCurrency,
				tt.exchangeRate,
				tt.exchangeDate,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, conversion)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conversion)
				assert.Equal(t, tt.transaction.ID, conversion.TransactionID)
			}
		})
	}
}
