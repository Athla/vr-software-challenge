package service

import (
	"context"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/treasury"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/google/uuid"
)

type CurrencyService struct {
	treasuryClient treasury.Clienter
	txRepo         repository.TransactionRepository
}

type CurrencyServicer interface {
	ConvertTransaction(ctx context.Context, transactionID uuid.UUID, targetCurrency string) (*models.CurrencyConversion, error)
}

func NewCurrencyService(
	treasuryClient treasury.Clienter,
	txRepo repository.TransactionRepository,
) *CurrencyService {
	return &CurrencyService{
		treasuryClient: treasuryClient,
		txRepo:         txRepo,
	}
}

func (s *CurrencyService) ConvertTransaction(
	ctx context.Context,
	transactionID uuid.UUID,
	targetCurrency string,
) (*models.CurrencyConversion, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	tx, err := s.txRepo.GetById(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	sixMonthsAgo := tx.TransactionDate.AddDate(0, -6, 0)
	rates, err := s.treasuryClient.GetExchangeRatesInRange(ctx, targetCurrency, sixMonthsAgo, tx.TransactionDate)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, errors.ErrNoValidExchangeRate
	}

	// Use the most recent rate
	rate := rates[0]

	conversion, err := models.NewCurrencyConversion(tx, targetCurrency, rate.Rate, rate.EffectiveDate)
	if err != nil {
		return nil, err
	}

	if err := conversion.Validate(); err != nil {
		return nil, err
	}

	return conversion, nil
}
