package handlers

import (
	go_errors "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/charmbracelet/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionHandler handles HTTP requests for transactions.
type TransactionHandler struct {
	Repo     repository.TransactionRepository
	Producer messagery.Producerer
}

// CreateTransactionRequest represents the request body for creating a transaction.
type CreateTransactionRequest struct {
	Description     string          `json:"description"`
	TransactionDate string          `json:"transaction_date"`
	AmountUSD       decimal.Decimal `json:"amount_usd"`
}

// CreateTransactionResponse represents the response body for creating a transaction.
type CreateTransactionResponse struct {
	ID      uuid.UUID `json:"id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
}

// Create handles the creation of a new transaction.
func (h *TransactionHandler) Create(ctx *gin.Context) {
	var req CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorf("Invalid body request: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format."})
		return
	}

	date, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		log.Errorf("Invalid transaction date format: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction date format"})
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
		log.Errorf("Transaction validation failed due: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.Create(ctx.Request.Context(), tx); err != nil {
		log.Errorf("Unable to create transaction due: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction."})
		return
	}

	msg := &messagery.TransactionMessage{
		ID:              tx.ID,
		Description:     tx.Description,
		TransactionDate: tx.TransactionDate,
		AmountUSD:       tx.AmountUSD,
		CreatedAt:       tx.CreatedAt,
	}

	if err := h.Producer.PublishTransaction(ctx.Request.Context(), msg); err != nil {
		for i := 0; i < 3; i++ {
			if err := h.Producer.PublishTransaction(ctx.Request.Context(), msg); err == nil {
				break
			}
			log.Errorf("Retry %d: failed to publish transaction due: %s", i+1, err)
			time.Sleep(2 * time.Second)
		}
		log.Errorf("Unable to publish transaction due: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish transaction"})
		return
	}

	ctx.JSON(http.StatusCreated, CreateTransactionResponse{
		ID:      tx.ID,
		Status:  string(tx.Status),
		Message: "Transaction created successfully.",
	})
}

// GetByID handles fetching a transaction by ID.
func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	tx, err := h.Repo.GetById(c.Request.Context(), id)
	if err != nil {
		if go_errors.Is(err, errors.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// UpdateStatus handles updating the status of a transaction.
func (h *TransactionHandler) UpdateStatus(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var req map[string]string
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body request"})
		return
	}

	status := models.TransactionStatus(req["status"])
	if err := h.Repo.UpdateStatus(ctx.Request.Context(), id, status); err != nil {
		if go_errors.Is(err, errors.ErrTransactionNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Transaction status updated successfully"})
}

// List handles listing transactions with pagination.
func (h *TransactionHandler) List(ctx *gin.Context) {
	limit := 10
	offset := 0

	if l := ctx.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := ctx.Query("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	transactions, err := h.Repo.List(ctx.Request.Context(), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list transactions"})
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}
