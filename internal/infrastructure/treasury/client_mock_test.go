package treasury

import (
	"context"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMockClient_GetExchangeRate(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Add the custom rate before testing
	client.AddMockRate("CAD", decimal.NewFromFloat(1.25), now.Add(-24*time.Hour))

	tests := []struct {
		name          string
		currency      string
		date          time.Time
		expectedRate  decimal.Decimal
		expectedError error
	}{
		{
			name:          "Valid EUR rate",
			currency:      "EUR",
			date:          now,
			expectedRate:  decimal.NewFromFloat(0.85),
			expectedError: nil,
		},
		{
			name:          "Invalid currency",
			currency:      "INVALID",
			date:          now,
			expectedRate:  decimal.Zero,
			expectedError: errors.ErrInvalidCurrency,
		},
		{
			name:          "Rate too old",
			currency:      "EUR",
			date:          now.AddDate(-1, 0, 0),
			expectedRate:  decimal.Zero,
			expectedError: errors.ErrNoValidExchangeRate,
		},
		{
			name:          "Custom mock rate",
			currency:      "CAD",
			date:          now,
			expectedRate:  decimal.NewFromFloat(1.25),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := client.GetExchangeRate(context.Background(), tt.currency, tt.date)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, rate)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rate)
				assert.Equal(t, tt.expectedRate.String(), rate.Rate.String())
				assert.Equal(t, tt.currency, rate.Currency)
			}
		})
	}
}

func TestMockClient_GetExchangeRatesInRange(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	tests := []struct {
		name          string
		currency      string
		startDate     time.Time
		endDate       time.Time
		expectedCount int
		expectedError error
	}{
		{
			name:          "Valid EUR rates",
			currency:      "EUR",
			startDate:     now.AddDate(0, -2, 0),
			endDate:       now,
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:          "Invalid currency",
			currency:      "INVALID",
			startDate:     now.AddDate(0, -1, 0),
			endDate:       now,
			expectedCount: 0,
			expectedError: errors.ErrInvalidCurrency,
		},
		{
			name:          "No rates in range",
			currency:      "EUR",
			startDate:     now.AddDate(-1, 0, 0),
			endDate:       now.AddDate(-1, 0, 0).AddDate(0, 1, 0),
			expectedCount: 0,
			expectedError: errors.ErrNoValidExchangeRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rates, err := client.GetExchangeRatesInRange(
				context.Background(),
				tt.currency,
				tt.startDate,
				tt.endDate,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, rates)
			} else {
				assert.NoError(t, err)
				assert.Len(t, rates, tt.expectedCount)
				for _, rate := range rates {
					assert.Equal(t, tt.currency, rate.Currency)
				}
			}
		})
	}
}
