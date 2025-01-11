package treasury

import (
	"context"
	"sync"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/shopspring/decimal"
)

type MockClient struct {
	rates map[string][]ExchangeRate
	mu    sync.Mutex // Add mutex for thread safety
}

func NewMockClient() *MockClient {
	return &MockClient{
		rates: map[string][]ExchangeRate{
			"EUR": {
				{
					Currency:      "EUR",
					Rate:          decimal.NewFromFloat(0.85),
					EffectiveDate: time.Now().AddDate(0, 0, -1), // Yesterday
				},
				{
					Currency:      "EUR",
					Rate:          decimal.NewFromFloat(0.84),
					EffectiveDate: time.Now().AddDate(0, -1, 0), // Last month
				},
			},
			"GBP": {
				{
					Currency:      "GBP",
					Rate:          decimal.NewFromFloat(0.73),
					EffectiveDate: time.Now().AddDate(0, 0, -1),
				},
			},
			"JPY": {
				{
					Currency:      "JPY",
					Rate:          decimal.NewFromFloat(110.0),
					EffectiveDate: time.Now().AddDate(0, 0, -1),
				},
			},
		},
	}
}

func (m *MockClient) GetExchangeRate(ctx context.Context, currency string, date time.Time) (*ExchangeRate, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rates, exists := m.rates[currency]
	if !exists {
		return nil, errors.ErrInvalidCurrency
	}

	var mostRecentRate *ExchangeRate
	for _, rate := range rates {
		if !rate.EffectiveDate.After(date) {
			if mostRecentRate == nil || rate.EffectiveDate.After(mostRecentRate.EffectiveDate) {
				rateCopy := rate
				mostRecentRate = &rateCopy
			}
		}
	}

	if mostRecentRate == nil {
		return nil, errors.ErrNoValidExchangeRate
	}

	sixMonthsAgo := date.AddDate(0, -6, 0)
	if mostRecentRate.EffectiveDate.Before(sixMonthsAgo) {
		return nil, errors.ErrNoValidExchangeRate
	}

	return mostRecentRate, nil
}

func (m *MockClient) GetExchangeRatesInRange(ctx context.Context, currency string, startDate, endDate time.Time) ([]ExchangeRate, error) {
	rates, exists := m.rates[currency]
	if !exists {
		return nil, errors.ErrInvalidCurrency
	}

	var result []ExchangeRate
	for _, rate := range rates {
		if !rate.EffectiveDate.Before(startDate) && !rate.EffectiveDate.After(endDate) {
			result = append(result, rate)
		}
	}

	if len(result) == 0 {
		return nil, errors.ErrNoValidExchangeRate
	}

	return result, nil
}

func (m *MockClient) AddMockRate(currency string, rate decimal.Decimal, date time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newRate := ExchangeRate{
		Currency:      currency,
		Rate:          rate,
		EffectiveDate: date,
	}

	if m.rates == nil {
		m.rates = make(map[string][]ExchangeRate)
	}

	m.rates[currency] = append(m.rates[currency], newRate)
}
