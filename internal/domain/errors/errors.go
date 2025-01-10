package errors

import "errors"

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrDescriptionTooLong     = errors.New("description exceeds 50 characters")
	ErrInvalidAmount          = errors.New("amount must be greater than zero")
	ErrFutureDate             = errors.New("transaction date cannot be in the future")
	ErrKafkaConnection        = errors.New("kafka connection error")
	ErrMessageValidation      = errors.New("message validation failed")
	ErrDatabaseConnection     = errors.New("database connection error")
	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrDescriptionEmpty       = errors.New("description cannot be empty")
)
