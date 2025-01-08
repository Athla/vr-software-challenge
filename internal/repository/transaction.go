package repository

import (
	"context"
	"database/sql"
	go_errors "errors"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetById(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error
	List(ctx context.Context, limit, offset int) ([]models.Transaction, error)
}

type postgresTransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &postgresTransactionRepo{
		db: db,
	}
}

func (r *postgresTransactionRepo) Create(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, description, transaction_date, amount_usd, status
		) VALUES (
		$1, $2, $3, $4, $5
	) RETURNING created_at`

	if err := r.db.QueryRowContext(
		ctx,
		query,
		tx.ID,
		tx.Description,
		tx.TransactionDate,
		tx.AmountUSD,
		tx.Status,
	).Scan(&tx.CreatedAt); err != nil {
		log.Errorf("Unable to create transaction due: %v", err)
		return err
	}

	return nil
}

func (r *postgresTransactionRepo) GetById(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `
	SELECT id, description, transaction_date, amount_usd, created_at, processed_at, status
	FROM transactions
	WHERE id = $1
	`

	tx := &models.Transaction{}
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.Description,
		&tx.TransactionDate,
		&tx.AmountUSD,
		&tx.CreatedAt,
		&tx.ProcessedAt,
		&tx.Status,
	); err != nil {
		if go_errors.Is(err, sql.ErrNoRows) {
			log.Error(errors.ErrTransactionNotFound)
			return nil, errors.ErrTransactionNotFound
		}
		log.Errorf("Unable to fetch transaction due: %v", err)
		return nil, err
	}

	return tx, nil
}

func (r *postgresTransactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error {
	query := `
		UPDATE transactions
		SET status = $1,
			processed_at = CASE
				WHEN $1 = 'COMPLETED' THEN CURRENT_TIMESTAMP
				ELSE processed_at
			END
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		log.Errorf("Unable to update transaction status due: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Errorf("Unable to check number of rows affected due: %v", err)
		return err
	}
	if rows == 0 {
		return errors.ErrTransactionNotFound
	}

	return nil
}

func (r *postgresTransactionRepo) List(ctx context.Context, limit, offset int) ([]models.Transaction, error) {
	query := `
        SELECT id, description, transaction_date, amount_usd, 
               created_at, processed_at, status
        FROM transactions 
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Errorf("Unable to query transactions due: %s", err)
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.Description,
			&tx.TransactionDate,
			&tx.AmountUSD,
			&tx.CreatedAt,
			&tx.ProcessedAt,
			&tx.Status,
		)
		if err != nil {
			log.Errorf("Unable to scan transaction due: %s", err)
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		log.Errorf("Unable to iterate over transactions due: %s", err)
		return nil, err
	}

	return transactions, nil
}
