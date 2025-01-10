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

type TransactionHandler struct {
	Repo     repository.TransactionRepository
	Producer messagery.Producerer
}

// CreateTransactionRequest represents the request body for creating a transaction.
// @Description Transaction creation request
type CreateTransactionRequest struct {
	// Description of the purchase
	// @Example "Office Supplies"
	Description string `json:"description" binding:"required" example:"Office Supplies"`

	// Date of the transaction in YYYY-MM-DD format
	// @Example "2024-01-10"
	TransactionDate string `json:"transaction_date" binding:"required" example:"2024-01-10"`

	// Amount in USD with up to 2 decimal places
	// @Example 123.45
	AmountUSD decimal.Decimal `json:"amount_usd" binding:"required" example:"123.45"`
}

// CreateTransactionResponse represents the response for a transaction creation.
// @Description Transaction creation response
type CreateTransactionResponse struct {
	// Unique identifier for the transaction
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Current status of the transaction
	// @Example "PENDING"
	Status string `json:"status" example:"PENDING"`

	// Response message
	// @Example "Transaction created successfully."
	Message string `json:"message" example:"Transaction created successfully."`
}

// @Summary Create a new transaction
// @Description Create a new purchase transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction Details"
// @Success 201 {object} CreateTransactionResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /transactions [post]
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
		AmountUSD:       req.AmountUSD.Round(2),
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

	var publishErr error
	for i := 0; i < 3; i++ {
		if err := h.Producer.PublishTransaction(ctx.Request.Context(), msg); err == nil {
			publishErr = nil
			break
		} else {
			publishErr = err
			log.Errorf("Retry %d: failed to publish transaction due: %s", i+1, err)
			time.Sleep(2 * time.Second)
		}
	}

	if publishErr != nil {
		log.Errorf("Failed to publish transaction after retries: %s", publishErr)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to processes transaction"})
		return
	}

	ctx.JSON(http.StatusCreated, CreateTransactionResponse{
		ID:      tx.ID,
		Status:  string(tx.Status),
		Message: "Transaction created successfully.",
	})
}

// @Summary Get a transaction by ID
// @Description Get a transaction's details by its ID
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} models.Transaction
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /transactions/{id} [get]
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

// @Summary Update transaction status
// @Description Update the status of a transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param status body map[string]string true "Status Update"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /transactions/{id}/status [patch]
func (h *TransactionHandler) UpdateStatus(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID format"})
		return
	}

	var req map[string]string
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	status := models.TransactionStatus(req["status"])
	if err := h.Repo.UpdateStatus(ctx.Request.Context(), id, status); err != nil {
		if go_errors.Is(err, errors.ErrTransactionNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		log.Errorf("Failed to update transaction status: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Transaction status updated successfully"})
}

// @Summary List transactions
// @Description Get a list of transactions with pagination
// @Tags transactions
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Transaction
// @Failure 500 {object} gin.H
// @Router /transactions [get]
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
