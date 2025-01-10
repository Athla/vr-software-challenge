package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/internal/api/handlers"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	repo := repository.NewMockTransactionRepository()
	producer := messagery.NewMockProducer()
	handler := handlers.TransactionHandler{
		Repo:     repo,
		Producer: producer,
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/transactions", handler.Create)

	reqBody := handlers.CreateTransactionRequest{
		Description:     "Test Transaction",
		TransactionDate: "2023-10-10",
		AmountUSD:       decimal.NewFromFloat(100.0),
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp handlers.CreateTransactionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resp.ID)
	assert.Equal(t, "PENDING", resp.Status)
	assert.Equal(t, "Transaction created successfully.", resp.Message)
}

func TestGetTransactionByID(t *testing.T) {
	repo := repository.NewMockTransactionRepository()
	producer := messagery.NewMockProducer()
	handler := handlers.TransactionHandler{
		Repo:     repo,
		Producer: producer,
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/transactions/:id", handler.GetByID)

	// Create a transaction to fetch
	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		Status:          models.StatusPending,
	}
	repo.Create(context.Background(), tx)

	req, _ := http.NewRequest("GET", "/transactions/"+tx.ID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var fetchedTx models.Transaction
	err := json.Unmarshal(rr.Body.Bytes(), &fetchedTx)
	assert.NoError(t, err)
	assert.Equal(t, tx.ID, fetchedTx.ID)
	assert.Equal(t, tx.Description, fetchedTx.Description)
	assert.Equal(t, tx.TransactionDate, fetchedTx.TransactionDate)
	assert.Equal(t, tx.AmountUSD, fetchedTx.AmountUSD)
	assert.Equal(t, tx.Status, fetchedTx.Status)
}

func TestUpdateTransactionStatus(t *testing.T) {
	repo := repository.NewMockTransactionRepository()
	producer := messagery.NewMockProducer()
	handler := handlers.TransactionHandler{
		Repo:     repo,
		Producer: producer,
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/transactions/:id/status", handler.UpdateStatus)

	// Create a transaction to update
	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		Status:          models.StatusPending,
	}
	repo.Create(context.Background(), tx)

	reqBody := map[string]string{"status": "COMPLETED"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", "/transactions/"+tx.ID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Transaction status updated successfully", resp["message"])

	updatedTx, _ := repo.GetById(context.Background(), tx.ID)
	assert.Equal(t, models.StatusCompleted, updatedTx.Status)
}

func TestListTransactions(t *testing.T) {
	repo := repository.NewMockTransactionRepository()
	producer := messagery.NewMockProducer()
	handler := handlers.TransactionHandler{
		Repo:     repo,
		Producer: producer,
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/transactions", handler.List)

	// Create some transactions to list
	for i := 0; i < 5; i++ {
		tx := &models.Transaction{
			ID:              uuid.New(),
			Description:     fmt.Sprintf("Test Transaction %d", i),
			TransactionDate: time.Now(),
			AmountUSD:       decimal.NewFromFloat(100.0),
			Status:          models.StatusPending,
		}
		repo.Create(context.Background(), tx)
	}

	req, _ := http.NewRequest("GET", "/transactions", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var transactions []models.Transaction
	err := json.Unmarshal(rr.Body.Bytes(), &transactions)
	assert.NoError(t, err)
	assert.Len(t, transactions, 5)
}
