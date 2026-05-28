package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/asdf57/bsw/internal/currency"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// GetUserStats godoc
// @Summary Get user spending statistics
// @Tags stats
// @Produce json
// @Param user query string true "User name"
// @Param currency query string false "Currency for generated stats" default(USD)
// @Success 200 {object} api.UserStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/stats/user [get]
func (h *Handlers) GetUserStats(c *gin.Context) {
	userName := strings.TrimSpace(c.Query("user"))
	if userName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is required"})
		return
	}

	targetCurrency := currency.NormalizeCurrencyCode(c.DefaultQuery("currency", "USD"))
	user, err := h.Db.GetUserFromName(userName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payments []dbmodels.PaymentDBEntry
	if err := h.Db.DB.Preload("Payer").Preload("Debtors").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query payments"})
		return
	}

	totalSpent := decimal.Zero
	spentOut := decimal.Zero
	paymentsIncluded := 0

	for _, payment := range payments {
		participated := userParticipatedInPayment(*user, payment)
		if !participated {
			continue
		}

		convertedAmount, err := currency.ConvertCurrencyWithCache(h.Db.DB, payment.Amount, payment.Currency, targetCurrency, payment.Date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to convert payment %d: %s", payment.ID, err.Error())})
			return
		}

		if payment.PayerID == user.ID {
			spentOut = spentOut.Add(convertedAmount)
		}

		if participated {
			participants := decimal.NewFromInt(int64(len(payment.Debtors) + 1))
			totalSpent = totalSpent.Add(convertedAmount.Div(participants))
			paymentsIncluded++
		}
	}

	c.JSON(http.StatusOK, apimodels.UserStatsResponse{
		User:             mappers.UserSummaryFromDB(*user),
		Currency:         targetCurrency,
		TotalSpent:       totalSpent,
		SpentOut:         spentOut,
		PaymentsIncluded: paymentsIncluded,
	})
}

func userParticipatedInPayment(user dbmodels.UserDBEntry, payment dbmodels.PaymentDBEntry) bool {
	if payment.PayerID == user.ID {
		return true
	}
	for _, debtor := range payment.Debtors {
		if debtor.ID == user.ID {
			return true
		}
	}
	return false
}
