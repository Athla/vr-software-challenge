package handlers

import (
	"net/http"

	"github.com/Athla/vr-software-challenge/internal/domain/errors"
	"github.com/Athla/vr-software-challenge/internal/service"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CurrencyHandler struct {
	CurrencyService service.CurrencyServicer
}

type ConversionResponse struct {
	TransactionID   string `json:"transaction_id"`
	Description     string `json:"description"`
	TransactionDate string `json:"transaction_date"`
	OriginalAmount  string `json:"original_amount_usd"`
	ExchangeRate    string `json:"exchange_rate"`
	ConvertedAmount string `json:"converted_amount"`
	TargetCurrency  string `json:"target_currency"`
	ExchangeDate    string `json:"exchange_date"`
}

// @Summary Convert transaction amount to different currency
// @Description Convert a transaction amount to a specified currency using Treasury exchange rates
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param currency query string true "Target Currency Code"
// @Success 200 {object} ConversionResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 422 {object} gin.H
// @Router /transactions/{id}/convert [get]
func (h *CurrencyHandler) ConvertCurrency(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Errorf("Invalid transaction id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id."})
		return
	}

	currency := ctx.Query("currency")
	if currency == "" {
		log.Error("No currency provided.")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "currency parameter is required."})
		return
	}

	conversion, err := h.CurrencyService.ConvertTransaction(ctx.Request.Context(), id, currency)
	if err != nil {
		switch err {
		case errors.ErrTransactionNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "transaction not found."})
		case errors.ErrNoValidExchangeRate:
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "No valid exchange rate found within six months of transaction date."})
		case errors.ErrInvalidCurrency:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency code"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert currency"})
		}
		return
	}

	ctx.JSON(http.StatusOK, ConversionResponse{
		TransactionID:   conversion.TransactionID.String(),
		Description:     conversion.Description,
		TransactionDate: conversion.TransactionDate.Format("2006-01-02"),
		OriginalAmount:  conversion.OriginalAmount.String(),
		ExchangeRate:    conversion.ExchangeRate.String(),
		ConvertedAmount: conversion.ConvertedAmount.String(),
		TargetCurrency:  conversion.TargetCurrency,
		ExchangeDate:    conversion.ExchangeDate.Format("2006-01-02"),
	})
}
