package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Athla/vr-software-challenge/internal/domain/models"
	"github.com/Athla/vr-software-challenge/internal/infrastructure/treasury"
	"github.com/Athla/vr-software-challenge/internal/repository"
	"github.com/Athla/vr-software-challenge/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestConvertCurrency(t *testing.T) {
	tests := []struct {
		name           string
		transactionID  string
		currency       string
		setupMock      func(*repository.MockTransactionRepository, *treasury.MockClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:          "Successful conversion",
			transactionID: "123e4567-e89b-12d3-a456-426614174000",
			currency:      "EUR",
			setupMock: func(repo *repository.MockTransactionRepository, tc *treasury.MockClient) {
				id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				tx := &models.Transaction{
					ID:              id,
					Description:     "Test Transaction",
					TransactionDate: time.Now().UTC(),
					AmountUSD:       decimal.NewFromFloat(100.00),
					Status:          models.StatusCompleted,
				}
				repo.Create(context.Background(), tx)
				tc.AddMockRate("EUR", decimal.NewFromFloat(0.85), time.Now().UTC().Add(-24*time.Hour))
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"converted_amount": "85",
				"target_currency":  "EUR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockTransactionRepository()
			mockTreasury := treasury.NewMockClient()

			if tt.setupMock != nil {
				tt.setupMock(mockRepo, mockTreasury)
			}

			service := service.NewCurrencyService(mockTreasury, mockRepo)
			handler := &CurrencyHandler{CurrencyService: service}

			router := gin.New()
			router.GET("/api/v1/transactions/:id/convert", handler.ConvertCurrency)

			url := fmt.Sprintf("/api/v1/transactions/%s/convert?currency=%s",
				tt.transactionID, tt.currency)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				for k, v := range tt.expectedBody {
					assert.Equal(t, v, response[k])
				}
			}
		})
	}
}
