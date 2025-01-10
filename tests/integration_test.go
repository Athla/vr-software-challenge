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
	"github.com/Athla/vr-software-challenge/internal/infrastructure/messagery"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var (
	dbHandler *sql.DB
	producer  messagery.Producerer
	cfg       *config.Config
)

func cleanup(t *testing.T) {
	_, err := dbHandler.Exec("TRUNCATE TABLE transactions, transaction_audit_logs CASCADE")
	assert.NoError(t, err, "Failed to cleanup test data")
}

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

	producer, err = messagery.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
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

	router.POST("/api/v1/transactions", transactionHandler.Create)
	router.GET("/api/v1/transactions/:id", transactionHandler.GetByID)
	router.PATCH("/api/v1/transactions/:id/status", transactionHandler.UpdateStatus)
	router.GET("/api/v1/transactions", transactionHandler.List)

	return router
}

func TestCreateTransactionIntegration(t *testing.T) {
	cleanup(t)
	router := setupRouter()
	repo := repository.NewTransactionRepository(dbHandler)

	description := "Test Transaction"
	amount := decimal.NewFromFloat(100.0)
	dateStr := time.Now().Format("2006-01-02")

	reqBody := handlers.CreateTransactionRequest{
		Description:     description,
		TransactionDate: dateStr,
		AmountUSD:       amount,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp handlers.CreateTransactionResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resp.ID)
	assert.Equal(t, string(models.StatusPending), resp.Status)

	// Verify the transaction in database
	tx, err := repo.GetById(context.Background(), resp.ID)
	assert.NoError(t, err)
	assert.Equal(t, description, tx.Description)
	assert.Equal(t, amount.String(), tx.AmountUSD.String())
	assert.Equal(t, models.StatusPending, tx.Status)
	assert.NotNil(t, tx.CreatedAt)
	assert.Nil(t, tx.ProcessedAt)
}

func TestGetTransactionByIDIntegration(t *testing.T) {
	cleanup(t)
	router := setupRouter()

	// Create a transaction with UTC time
	now := time.Now().UTC()
	amount := decimal.NewFromFloat(100.0).Round(2) // Ensure 2 decimal places

	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: now.Truncate(24 * time.Hour), // Truncate to day precision
		AmountUSD:       amount,
		Status:          models.StatusPending,
	}
	err := repository.NewTransactionRepository(dbHandler).Create(context.Background(), tx)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/transactions/"+tx.ID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var fetchedTx models.Transaction
	err = json.Unmarshal(rr.Body.Bytes(), &fetchedTx)
	assert.NoError(t, err)
	assert.Equal(t, tx.ID, fetchedTx.ID)
	assert.Equal(t, tx.Description, fetchedTx.Description)

	// Compare dates after truncating to day precision and converting to UTC
	expectedDate := tx.TransactionDate.UTC().Truncate(24 * time.Hour)
	actualDate := fetchedTx.TransactionDate.UTC().Truncate(24 * time.Hour)
	assert.Equal(t, expectedDate, actualDate)

	// Compare amounts after ensuring same precision
	expectedAmount := tx.AmountUSD.Round(2)
	actualAmount := fetchedTx.AmountUSD.Round(2)
	assert.Equal(t, expectedAmount.String(), actualAmount.String())

	assert.Equal(t, tx.Status, fetchedTx.Status)
}

func TestListTransactionsIntegration(t *testing.T) {
	cleanup(t)
	router := setupRouter()
	repo := repository.NewTransactionRepository(dbHandler)

	expectedTxs := make([]*models.Transaction, 5)
	for i := 0; i < 5; i++ {
		tx := &models.Transaction{
			ID:              uuid.New(),
			Description:     fmt.Sprintf("Test Transaction %d", i),
			TransactionDate: time.Now().AddDate(0, 0, -i), // Different dates
			AmountUSD:       decimal.NewFromFloat(100.0 + float64(i)),
			Status:          models.StatusPending,
		}
		err := repo.Create(context.Background(), tx)
		assert.NoError(t, err)
		expectedTxs[i] = tx
		time.Sleep(100 * time.Millisecond) // Ensure different created_at times
	}

	req, _ := http.NewRequest("GET", "/api/v1/transactions?limit=3&offset=0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var transactions []models.Transaction
	err := json.Unmarshal(rr.Body.Bytes(), &transactions)
	assert.NoError(t, err)
	assert.Len(t, transactions, 3, "Should return exactly 3 transactions")

	for i := 0; i < len(transactions)-1; i++ {
		assert.True(t, transactions[i].CreatedAt.After(transactions[i+1].CreatedAt),
			"Transactions should be ordered by created_at DESC")
	}
}

func TestUpdateTransactionStatusIntegration(t *testing.T) {
	cleanup(t)
	router := setupRouter()

	// Create a transaction
	tx := &models.Transaction{
		ID:              uuid.New(),
		Description:     "Test Transaction",
		TransactionDate: time.Now(),
		AmountUSD:       decimal.NewFromFloat(100.0),
		Status:          models.StatusPending,
	}
	repo := repository.NewTransactionRepository(dbHandler)
	err := repo.Create(context.Background(), tx)
	assert.NoError(t, err)

	// Verify initial status
	initialTx, err := repo.GetById(context.Background(), tx.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusPending, initialTx.Status)

	// Update status
	reqBody := map[string]string{"status": "COMPLETED"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", "/api/v1/transactions/"+tx.ID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Add a small delay to allow for processing
	time.Sleep(1 * time.Second)

	// Verify final status
	updatedTx, err := repo.GetById(context.Background(), tx.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusCompleted, updatedTx.Status)
	assert.NotNil(t, updatedTx.ProcessedAt, "ProcessedAt should be set when status is COMPLETED")
}
