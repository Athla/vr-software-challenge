package models

import (
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CurrencyConversion struct {
	TransactionID   uuid.UUID
	Description     string
	TransactionDate time.Time
	OriginalAmount  decimal.Decimal
	ExchangeRate    decimal.Decimal
	ExchangeDate    time.Time
	TargetCurrency  string
	ConvertedAmount decimal.Decimal
}

func NewCurrencyConversion(
	tx *Transaction,
	targetCurrency string,
	exchangeRate decimal.Decimal,
	exchangeDate time.Time,
) (*CurrencyConversion, error) {
	if err := validateExchangeDate(tx.TransactionDate, exchangeDate); err != nil {
		return nil, err
	}

	convertedAmount := tx.AmountUSD.Mul(exchangeRate).Round(4)

	return &CurrencyConversion{
		TransactionID:   tx.ID,
		Description:     tx.Description,
		TransactionDate: tx.TransactionDate,
		OriginalAmount:  tx.AmountUSD,
		ExchangeRate:    exchangeRate.Round(4),
		ExchangeDate:    exchangeDate,
		TargetCurrency:  targetCurrency,
		ConvertedAmount: convertedAmount,
	}, nil
}

func validateExchangeDate(transactionDate, exchangeDate time.Time) error {
	if exchangeDate.After(transactionDate) {
		return errors.ErrNoValidExchangeRate
	}

	sixMonthAgo := transactionDate.AddDate(0, -6, 0)
	if exchangeDate.Before(sixMonthAgo) {
		return errors.ErrNoValidExchangeRate
	}

	return nil
}

func (c *CurrencyConversion) Validate() error {
	if c.TransactionID == uuid.Nil {
		return errors.ErrTransactionNotFound
	}

	if c.TargetCurrency == "" {
		return errors.ErrInvalidCurrency
	}

	if c.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		return errors.ErrNoValidExchangeRate
	}

	if err := validateExchangeDate(c.TransactionDate, c.ExchangeDate); err != nil {
		return err
	}

	return nil
}
