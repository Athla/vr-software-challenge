package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/Athla/vr-software-challenge/internal/api/handlers"
	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/database"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/kafka"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var (
	dbHandler *sql.DB
	producer  kafka.Producerer
	cfg       *config.Config
)

func TestMain(m *testing.M) {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Printf("Unable to load config: %v\n", err)
		os.Exit(1)
	}

	dbHandler, err = database.NewConnection(cfg.Database)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	producer, err = kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.TransactionTopic)
	if err != nil {
		fmt.Printf("Unable to create Kafka producer: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	dbHandler.Close()
	producer.Close()

	os.Exit(code)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	transactionHandler := handlers.TransactionHandler{
		Repo:     repository.NewTransactionRepository(dbHandler),
		Producer: producer,
	}

	router.POST("/transactions", transactionHandler.Create)
	router.GET("/transactions/:id", transactionHandler.GetByID)
	router.PATCH("/transactions/:id/status", transactionHandler.UpdateStatus)
	router.GET("/transactions", transactionHandler.List)

	return router
}

func TestCreateTransactionIntegration(t *testing.T) {
	router := setupRouter()

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

func TestGetTransactionByIDIntegration(t *testing.T) {
	router := setupRouter()

	// Create a transaction to fetch
	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		Status:          models.StatusPending,
	}
	err := repository.NewTransactionRepository(dbHandler).Create(context.Background(), tx)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/transactions/"+tx.ID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var fetchedTx models.Transaction
	err = json.Unmarshal(rr.Body.Bytes(), &fetchedTx)
	assert.NoError(t, err)
	assert.Equal(t, tx.ID, fetchedTx.ID)
	assert.Equal(t, tx.Description, fetchedTx.Description)
	assert.Equal(t, tx.TransactionDate, fetchedTx.TransactionDate)
	assert.Equal(t, tx.AmountUSD, fetchedTx.AmountUSD)
	assert.Equal(t, tx.Status, fetchedTx.Status)
}

func TestUpdateTransactionStatusIntegration(t *testing.T) {
	router := setupRouter()

	// Create a transaction to update
	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		Status:          models.StatusPending,
	}
	err := repository.NewTransactionRepository(dbHandler).Create(context.Background(), tx)
	assert.NoError(t, err)

	reqBody := map[string]string{"status": "COMPLETED"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", "/transactions/"+tx.ID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Transaction status updated successfully", resp["message"])

	updatedTx, err := repository.NewTransactionRepository(dbHandler).GetById(context.Background(), tx.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusCompleted, updatedTx.Status)
}

func TestListTransactionsIntegration(t *testing.T) {
	router := setupRouter()

	// Create some transactions to list
	for i := 0; i < 5; i++ {
		tx := &models.Transaction{
			ID:              uuid.New(),
			Description:     fmt.Sprintf("Test Transaction %d", i),
			TransactionDate: time.Now(),
			AmountUSD:       decimal.NewFromFloat(100.0),
			Status:          models.StatusPending,
		}
		err := repository.NewTransactionRepository(dbHandler).Create(context.Background(), tx)
		assert.NoError(t, err)
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
