package treasury

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/charmbracelet/log"
	"github.com/shopspring/decimal"
)

const (
	baseURL      = "https://api.fiscaldata.treasury.gov/services/v1/accounting/od/rates_of_exchange"
	rateEndpoint = "/rates_of_exchange"
)

type Clienter interface {
	GetExchangeRate(ctx context.Context, currency string, date time.Time) (*ExchangeRate, error)
	GetExchangeRatesInRange(ctx context.Context, currency string, startDate, endDate time.Time) ([]ExchangeRate, error)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithHttpClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ExchangeRate struct {
	Currency      string          `json:"currency"`
	Rate          decimal.Decimal `json:"exchange_rate"`
	EffectiveDate time.Time       `json:"effective_date"`
}

type ExchangeRateResponse struct {
	Data []struct {
		CountryCode   string `json:"country_currency_desc"`
		ExchangeRate  string `json:"exchange_rate"`
		EffectiveDate string `json:"record_date"`
	} `json:"data"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

func (c *Client) GetExchangeRate(ctx context.Context, currency string, date time.Time) (*ExchangeRate, error) {
	params := url.Values{}
	params.Add("fields", "country_currency_desc,exchange_rate,record_date")
	params.Add("filter", fmt.Sprintf("country_currency_desc:eq:%s,record_date:lte:%s",
		currency, date.Format("2006-01-02")))
	params.Add("sort", "-record_date")
	params.Add("limit", "1")

	reqUrl := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		log.Errorf("Failed to create request: %s", err)
		return nil, errors.ErrTreasuryAPIError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Errorf("Failed to execute request: %v", err)
		return nil, fmt.Errorf("%w: failed to execute request", errors.ErrTreasuryAPIError)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Treasure API error: %s - status %d", string(body), resp.StatusCode)
		return nil, fmt.Errorf("%w: unexpected status code %d", errors.ErrTreasuryAPIError, resp.StatusCode)
	}
	var apiResp ExchangeRateResponse
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Errorf("Failed to decode response: %s", err)
		return nil, fmt.Errorf("%w: Failed to decode response.", errors.ErrTreasuryAPIError)
	}

	if len(apiResp.Data) == 0 {
		return nil, errors.ErrNoValidExchangeRate
	}

	rate, err := decimal.NewFromString(apiResp.Data[0].ExchangeRate)
	if err != nil {
		log.Errorf("Failed to parse exchange rate: %v", err)
		return nil, fmt.Errorf("%w: failed to parse exchange rate", errors.ErrTreasuryAPIError)
	}

	effectiveDate, err := time.Parse("2006-01-02", apiResp.Data[0].EffectiveDate)
	if err != nil {
		log.Errorf("Failed to parse effective date: %v", err)
		return nil, fmt.Errorf("%w: failed to parse effective date", errors.ErrTreasuryAPIError)
	}

	return &ExchangeRate{
		Currency:      currency,
		Rate:          rate,
		EffectiveDate: effectiveDate,
	}, nil
}

func (c *Client) GetExchangeRatesInRange(ctx context.Context, currency string, startDate, endDate time.Time) ([]ExchangeRate, error) {
	params := url.Values{}
	params.Add("fields", "country_currency_desc,exchange_rate,record_date")
	params.Add("filter", fmt.Sprintf("country_currency_desc:eq:%s,record_date:gte:%s,record_date:lte:%s",
		currency,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")))
	params.Add("sort", "-record_date")

	reqURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", errors.ErrTreasuryAPIError)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to execute request", errors.ErrTreasuryAPIError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Treasury API error: %s, status: %d", string(body), resp.StatusCode)
		return nil, fmt.Errorf("%w: unexpected status code %d", errors.ErrTreasuryAPIError, resp.StatusCode)
	}

	var apiResp ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response", errors.ErrTreasuryAPIError)
	}

	rates := make([]ExchangeRate, 0, len(apiResp.Data))
	for _, data := range apiResp.Data {
		rate, err := decimal.NewFromString(data.ExchangeRate)
		if err != nil {
			log.Warnf("Skipping invalid rate for %s: %v", data.CountryCode, err)
			continue
		}

		effectiveDate, err := time.Parse("2006-01-02", data.EffectiveDate)
		if err != nil {
			log.Warnf("Skipping invalid date for %s: %v", data.CountryCode, err)
			continue
		}

		rates = append(rates, ExchangeRate{
			Currency:      currency,
			Rate:          rate,
			EffectiveDate: effectiveDate,
		})
	}

	return rates, nil
}
