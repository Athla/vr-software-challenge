package handlers

import (
	"encoding/json"
	go_errors "errors"
	"net/http"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/kafka"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/charmbracelet/log"

	// Change to gin.
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	repo     repository.TransactionRepository
	producer *kafka.Producer
}

type CreateTransactionRequest struct {
	Description     string          `json:"description"`
	TransactionDate string          `json:"transaction_date"`
	AmountUSD       decimal.Decimal `json:"amount_usd"`
}

type CreateTransactionResponse struct {
	ID      uuid.UUID `json:"id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid transaction date format")
		return
	}

	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     req.Description,
		TransactionDate: date,
		AmountUSD:       req.AmountUSD,
		Status:          models.StatusPending,
	}

	if err := tx.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Store transaction
	if err := h.repo.Create(r.Context(), tx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	// Publish to Kafka
	msg := &kafka.TransactionMessage{
		ID:              tx.ID,
		Description:     tx.Description,
		TransactionDate: tx.TransactionDate,
		AmountUSD:       tx.AmountUSD,
		CreatedAt:       tx.CreatedAt,
	}

	if err := h.producer.PublishTransaction(r.Context(), msg); err != nil {
		// TODO: Implement retry logic.
		log.Errorf("failed to publish transaction", "error", err)
	}

	respond(w, http.StatusCreated, CreateTransactionResponse{
		ID:      tx.ID,
		Status:  string(tx.Status),
		Message: "Transaction created successfully",
	})
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	tx, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		if go_errors.Is(err, errors.ErrTransactionNotFound) {
			respondError(w, http.StatusNotFound, "Transaction not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to fetch transaction")
		return
	}

	respond(w, http.StatusOK, tx)
}
