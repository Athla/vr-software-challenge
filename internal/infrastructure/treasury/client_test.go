package treasury

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetExchangeRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ExchangeRateResponse{
			Data: []struct {
				CountryCode   string `json:"country_currency_desc"`
				ExchangeRate  string `json:"exchange_rate"`
				EffectiveDate string `json:"record_date"`
			}{
				{
					CountryCode:   "Euro Zone-Euro",
					ExchangeRate:  "0.85",
					EffectiveDate: "2024-01-01",
				},
			},
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 1,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHttpClient(server.Client()),
	)

	rate, err := client.GetExchangeRate(context.Background(), "Euro Zone-Euro", time.Now())

	t.Log(rate)

	assert.NoError(t, err)
	assert.NotNil(t, rate)
	assert.Equal(t, "Euro Zone-Euro", rate.Currency)
	assert.Equal(t, decimal.NewFromFloat(0.85), rate.Rate)
	assert.Equal(t, "2024-01-01", rate.EffectiveDate.Format("2006-01-02"))
}

func TestGetExchangeRatesInRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ExchangeRateResponse{
			Data: []struct {
				CountryCode   string `json:"country_currency_desc"`
				ExchangeRate  string `json:"exchange_rate"`
				EffectiveDate string `json:"record_date"`
			}{
				{
					CountryCode:   "Euro Zone-Euro",
					ExchangeRate:  "0.85",
					EffectiveDate: "2024-01-01",
				},
				{
					CountryCode:   "Euro Zone-Euro",
					ExchangeRate:  "0.84",
					EffectiveDate: "2023-12-01",
				},
			},
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 2,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHttpClient(server.Client()),
	)

	startDate := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	rates, err := client.GetExchangeRatesInRange(context.Background(), "Euro Zone-Euro", startDate, endDate)

	t.Log(rates)
	assert.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.Equal(t, decimal.NewFromFloat(0.85), rates[0].Rate)
	assert.Equal(t, decimal.NewFromFloat(0.84), rates[1].Rate)
}
