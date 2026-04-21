package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/asdf57/bsw/internal/currency"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/gin-gonic/gin"
)

// GetExchangeRate godoc
// @Summary Get exchange rate between two currencies
// @Description Returns the exchange rate for `from` -> `to` on a given date (defaults to today if date is omitted).
// @Tags admin
// @Produce json
// @Param from query string true "Base currency code (e.g. USD)"
// @Param to query string true "Target currency code (e.g. EUR)"
// @Param date query string false "Date in YYYY-MM-DD format"
// @Success 200 {object} api.Exchange
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/exchange-rate [get]
func (h *Handlers) GetExchangeRate(c *gin.Context) {
	fromCurrency := strings.ToUpper(c.Query("from"))
	toCurrency := strings.ToUpper(c.Query("to"))
	dateQuery := c.Query("date")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing from or to query parameter"})
		return
	}

	at := time.Now().UTC()
	if dateQuery != "" {
		parsedDate, err := time.Parse("2006-01-02", dateQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date; expected YYYY-MM-DD"})
			return
		}
		at = parsedDate
	}

	if fromCurrency == toCurrency {
		c.JSON(http.StatusOK, apimodels.Exchange{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         1.0,
			Date:         at,
		})
		return
	}

	rate, err := currency.GetExchangeRate(fromCurrency, toCurrency, at)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch exchange rate: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, rate)
}
