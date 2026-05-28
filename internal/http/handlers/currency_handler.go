package handlers

import (
	"fmt"
	"net/http"
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
	rawFromCurrency := c.Query("from")
	rawToCurrency := c.Query("to")
	dateQuery := c.Query("date")

	if rawFromCurrency == "" || rawToCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing from or to query parameter"})
		return
	}
	fromCurrency := currency.NormalizeCurrencyCode(rawFromCurrency)
	toCurrency := currency.NormalizeCurrencyCode(rawToCurrency)

	at := time.Now().UTC()
	if dateQuery != "" {
		parsedDate, err := time.Parse("2006-01-02", dateQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date; expected YYYY-MM-DD"})
			return
		}
		at = parsedDate
	}

	rate, err := currency.CacheExchangeRate(h.Db.DB, apimodels.Exchange{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         at,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch exchange rate: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, apimodels.Exchange{
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Date:         rate.Date,
	})
}

// GetExchangeRatesFromDB godoc
// @Summary Get all exchange rates stored in the DB
// @Description Returns the exchange rates stored in the DB
// @Tags admin
// @Produce json
// @Success 200 {array} api.Exchange
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/exchange-rates [get]
func (h *Handlers) GetExchangeRatesFromDB(c *gin.Context) {
	exchangeRateDbEntries, err := h.Db.GetExchangeRatesFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to load exchange rates: %s", err.Error())})
		return
	}

	exchangeRates := make([]apimodels.Exchange, 0, len(exchangeRateDbEntries))
	for _, entry := range exchangeRateDbEntries {
		exchangeRates = append(exchangeRates, apimodels.Exchange{
			FromCurrency: entry.FromCurrency,
			ToCurrency:   entry.ToCurrency,
			Date:         entry.Date,
			Rate:         entry.Rate,
		})
	}

	c.JSON(http.StatusOK, exchangeRates)
}
