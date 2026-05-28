package currency

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/asdf57/bsw/internal/db"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func NormalizeExchangeDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func NormalizeCurrencyCode(code string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(code))
	if cleaned == "" {
		return "USD"
	}
	return cleaned
}

func ConvertCurrency(amount decimal.Decimal, exchangeRate apimodels.Exchange) (decimal.Decimal, error) {
	exchangeRate.FromCurrency = NormalizeCurrencyCode(exchangeRate.FromCurrency)
	exchangeRate.ToCurrency = NormalizeCurrencyCode(exchangeRate.ToCurrency)
	rate := exchangeRate.Rate

	if exchangeRate.FromCurrency == exchangeRate.ToCurrency {
		return amount, nil
	}

	if rate == 0 {
		return decimal.Zero, fmt.Errorf("missing exchange rate for %s to %s; use ConvertCurrencyWithCache", exchangeRate.FromCurrency, exchangeRate.ToCurrency)
	}

	return amount.Mul(decimal.NewFromFloat(rate)), nil
}

func ConvertCurrencyWithCache(tx *gorm.DB, amount decimal.Decimal, fromCurrency string, toCurrency string, date time.Time) (decimal.Decimal, error) {
	fromCurrency = NormalizeCurrencyCode(fromCurrency)
	toCurrency = NormalizeCurrencyCode(toCurrency)
	if fromCurrency == toCurrency {
		return amount, nil
	}

	exchange, err := CacheExchangeRate(tx, apimodels.Exchange{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         NormalizeExchangeDate(date),
	})
	if err != nil {
		return decimal.Zero, err
	}

	return ConvertCurrency(amount, apimodels.Exchange{
		FromCurrency: exchange.FromCurrency,
		ToCurrency:   exchange.ToCurrency,
		Date:         exchange.Date,
		Rate:         exchange.Rate,
	})
}

func GetExchangeRate(fromCurrency, toCurrency string, time time.Time) (apimodels.Exchange, error) {
	fromCurrency = NormalizeCurrencyCode(fromCurrency)
	toCurrency = NormalizeCurrencyCode(toCurrency)
	normalizedTime := NormalizeExchangeDate(time)
	dateStr := normalizedTime.Format("2006-01-02")

	if fromCurrency == toCurrency {
		return apimodels.Exchange{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         1.0,
			Date:         normalizedTime,
		}, nil
	}

	url := fmt.Sprintf("https://api.frankfurter.dev/v2/rates?base=%s&quotes=%s&date=%s", fromCurrency, toCurrency, dateStr)

	log.Printf("Hitting URL: %s", url)

	res, err := http.Get(url)
	if err != nil {
		fmt.Print("failed to obtain exchange rate")
		return apimodels.Exchange{}, fmt.Errorf("failed to obtain exchange rate: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return apimodels.Exchange{}, fmt.Errorf("failed to retrieve the exchange rate from API: %s", res.Status)
	}

	var queryRes []struct {
		Date  string  `json:"date"`
		Base  string  `json:"base"`
		Quote string  `json:"quote"`
		Rate  float64 `json:"rate"`
	}

	if err := json.NewDecoder(res.Body).Decode(&queryRes); err != nil {
		return apimodels.Exchange{}, fmt.Errorf("failed to coerce exchange rate data: %s", err)
	}

	log.Printf("Got exchange rate res: %+v", queryRes)

	if len(queryRes) != 1 {
		return apimodels.Exchange{}, fmt.Errorf("unexpected length from exchange rate query: %d", len(queryRes))
	}

	currencyExchange := apimodels.Exchange{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         queryRes[0].Rate,
		Date:         normalizedTime,
	}

	return currencyExchange, nil
}

func GetExchangeRateDBEntry(fromCurrency, toCurrency string, time time.Time) (dbmodels.ExchangeDBEntry, error) {
	normalizedTime := NormalizeExchangeDate(time)
	exchangeData, err := GetExchangeRate(fromCurrency, toCurrency, normalizedTime)
	if err != nil {
		return dbmodels.ExchangeDBEntry{}, err
	}

	return dbmodels.ExchangeDBEntry{
		FromCurrency: exchangeData.FromCurrency,
		ToCurrency:   exchangeData.ToCurrency,
		Rate:         exchangeData.Rate,
		Date:         normalizedTime,
	}, nil
}

func CacheExchangeRate(tx *gorm.DB, exchangeRateReq apimodels.Exchange) (dbmodels.ExchangeDBEntry, error) {
	var exchangeRateDbEntry dbmodels.ExchangeDBEntry

	exchangeRateReq.FromCurrency = NormalizeCurrencyCode(exchangeRateReq.FromCurrency)
	exchangeRateReq.ToCurrency = NormalizeCurrencyCode(exchangeRateReq.ToCurrency)
	exchangeRateReq.Date = NormalizeExchangeDate(exchangeRateReq.Date)

	dbQuery := dbmodels.ExchangeDBEntry{
		FromCurrency: exchangeRateReq.FromCurrency,
		ToCurrency:   exchangeRateReq.ToCurrency,
		Date:         exchangeRateReq.Date,
	}

	if err := tx.Where(&dbQuery).First(&exchangeRateDbEntry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Caching exchange rate entry with properties: %+v", exchangeRateReq)
			exchangeRateDbEntry, err = GetExchangeRateDBEntry(exchangeRateReq.FromCurrency, exchangeRateReq.ToCurrency, exchangeRateReq.Date)
			if err != nil {
				return dbmodels.ExchangeDBEntry{}, fmt.Errorf("failed to obtain exchange rate data: %w", err)
			}

			if err := tx.Create(&exchangeRateDbEntry).Error; err != nil {
				if db.IsUniqueConstraintError(err) {
					// conflict while caching, try re-reading instead!
					if err := tx.Where(&dbQuery).First(&exchangeRateDbEntry).Error; err != nil {
						return dbmodels.ExchangeDBEntry{}, fmt.Errorf("re-read of cached exchange rate entry failed: %w", err)
					}
				} else {
					return dbmodels.ExchangeDBEntry{}, fmt.Errorf("failed to cache exchange rate data: %w", err)
				}
			}
		} else {
			return dbmodels.ExchangeDBEntry{}, fmt.Errorf("failed to query for exchange rate data in DB: %w", err)
		}
	}

	return exchangeRateDbEntry, nil
}
