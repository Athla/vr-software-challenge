package models

import (
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCompleted  TransactionStatus = "COMPLETED"
	StatusFailed     TransactionStatus = "FAILED"
)

type Transaction struct {
	ID              uuid.UUID         `db:"id" json:"id"`
	Description     string            `db:"description" json:"description"`
	TransactionDate time.Time         `db:"transactiondate" json:"transactiondate"`
	AmountUSD       decimal.Decimal   `db:"amountusd" json:"amountusd"`
	CreatedAt       time.Time         `db:"createdat" json:"createdat"`
	ProcessedAt     *time.Time        `db:"processedat" json:"processedat"`
	Status          TransactionStatus `db:"status" json:"status"`
}

func (t *Transaction) Validate() error {
	if len(t.Description) > 50 {
		return errors.ErrDescriptionTooLong
	}

	if t.AmountUSD.LessThanOrEqual(decimal.Zero) {
		return errors.ErrInvalidAmount
	}

	if t.TransactionDate.After(time.Now()) {
		return errors.ErrFutureDate
	}

	return nil
}
