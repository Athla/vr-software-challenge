package messagery

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionMessage struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	Description     string          `db:"description" json:"description"`
	TransactionDate time.Time       `db:"transactiondate" json:"transactiondate"`
	AmountUSD       decimal.Decimal `db:"amountusd" json:"amountusd"`
	CreatedAt       time.Time       `db:"createdat" json:"createdat"`
}
