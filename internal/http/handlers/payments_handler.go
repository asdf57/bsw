package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/asdf57/bsw/internal/currency"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetPayment godoc
// @Summary Get a payment by ID
// @Tags payment
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} api.PaymentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/{id} [get]
func (h *Handlers) GetPayment(c *gin.Context) {
	id := c.Param("id")

	var p dbmodels.PaymentDBEntry
	if err := h.Db.DB.Preload("Payer").Preload("Debtors").Preload("Exchange").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponseFromDB(p))
}

// GetPayments godoc
// @Summary Get all payments
// @Tags payment
// @Produce json
// @Success 200 {array} api.PaymentResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/all [get]
func (h *Handlers) GetAllPayments(c *gin.Context) {
	var p []dbmodels.PaymentDBEntry

	if err := h.Db.DB.Preload("Payer").Preload("Debtors").Preload("Exchange").Find(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponsesFromDB(p))
}

// PostPayment godoc
// @Summary Create a new payment
// @Tags payment
// @Accept json
// @Produce json
// @Param payment body api.Payment true "Payment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment [post]
func (h *Handlers) PostPayment(c *gin.Context) {
	var payment apimodels.Payment

	if err := c.ShouldBind(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payerId, err := h.Db.GetUserIdFromName(payment.Payer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to obtain payer ID from DB: %s", err)})
		return
	}

	exchangeRateDbEntry, err := currency.GetExchangeRateDBEntry(payment.FromExchangeRate, payment.ToExchangeRate, payment.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	debtors, err := h.Db.GetUsersFromNames(payment.Debtors)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to obtain debtor data from DB: %s", err.Error())})
		return
	}

	debtorIDsByName := make(map[string]uint, len(debtors))
	for _, debtor := range debtors {
		debtorIDsByName[debtor.Name] = debtor.ID
	}

	for _, debtorName := range payment.Debtors {
		if _, ok := debtorIDsByName[debtorName]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no user with name %s could be found", debtorName)})
			return
		}
	}

	var paymentId uint

	err = h.Db.DB.Transaction(func(tx *gorm.DB) error {
		lookupKey := dbmodels.ExchangeDBEntry{
			FromCurrency: exchangeRateDbEntry.FromCurrency,
			ToCurrency:   exchangeRateDbEntry.ToCurrency,
			Date:         exchangeRateDbEntry.Date,
		}

		if err := tx.Where(&lookupKey).FirstOrCreate(&exchangeRateDbEntry).Error; err != nil {
			return fmt.Errorf("upsert exchange rate: %w", err)
		}

		parsedAmount := decimal.NewFromFloat(payment.Amount)

		record := dbmodels.PaymentDBEntry{
			Amount:      parsedAmount,
			Description: payment.Description,
			Date:        payment.Date,
			PayerID:     payerId,
			ExchangeID:  exchangeRateDbEntry.ID,
			Debtors:     debtors,
		}

		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create payment: %s", err.Error())
		}

		paymentId = record.ID

		// Now create associated debts (if specified)
		if len(payment.Debtors) > 0 {
			if payment.DebtMode != "equal" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "payment "})
				return fmt.Errorf("requested debt mode is either invalid or unsupported")
			}

			parsedNumDebtors := decimal.NewFromInt(int64(len(payment.Debtors)))
			amountOwedPerDebtor := decimal.NewFromFloat(payment.Amount).Div(parsedNumDebtors)

			// Users cannot owe debts to themselves -- fail fast, avoid extra DB operations!
			for _, debtor := range payment.Debtors {
				if debtor == payment.Payer {
					return fmt.Errorf("user %s cannot be indebted to themselves", debtor)
				}
			}

			for _, debtor := range payment.Debtors {
				debtEntry := dbmodels.DebtDBEntry{
					OwedByUserId: debtorIDsByName[debtor],
					OwedToUserId: payerId,
					Amount:       amountOwedPerDebtor,
					Currency:     payment.ToExchangeRate,
				}

				if err := tx.Create(&debtEntry).Error; err != nil {
					return fmt.Errorf("could not create debt entry in the db: %s", err.Error())
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create payment: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("payment created with id %d", paymentId)})
}

// DeletePayment godoc
// @Summary Delete a payment by ID
// @Tags payment
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/{id} [delete]
func (h *Handlers) DeletePayment(c *gin.Context) {
	id := c.Param("id")

	res := h.Db.DB.Delete(dbmodels.PaymentDBEntry{}, id)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create payment: %s", res.Error.Error())})
		return
	}

	if res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("payment with id %s does not exist!", id)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("deleted payment with id %s", id)})
}
