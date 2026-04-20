package currency

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/asdf57/bsw/internal/models"
)

func GetExchangeRate(fromCurrency, toCurrency string, time time.Time) (models.Exchange, error) {
	dateStr := time.Format("2006-01-02")

	if fromCurrency == toCurrency {
		return models.Exchange{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         1.0,
			Date:         time,
		}, nil
	}

	url := fmt.Sprintf("https://api.frankfurter.dev/v2/rates?base=%s&quotes=%s&date=%s", fromCurrency, toCurrency, dateStr)

	log.Printf("Hitting URL: %s", url)

	res, err := http.Get(url)
	if err != nil {
		fmt.Print("failed to obtain exchange rate")
		return models.Exchange{}, fmt.Errorf("failed to obtain exchange rate: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return models.Exchange{}, fmt.Errorf("failed to retrieve the exchange rate from API: %s", res.Status)
	}

	var queryRes []struct {
		Date  string  `json:"date"`
		Base  string  `json:"base"`
		Quote string  `json:"quote"`
		Rate  float64 `json:"rate"`
	}

	if err := json.NewDecoder(res.Body).Decode(&queryRes); err != nil {
		return models.Exchange{}, fmt.Errorf("failed to coerce exchange rate data: %s", err)
	}

	log.Printf("Got exchange rate res: %+v", queryRes)

	if len(queryRes) != 1 {
		return models.Exchange{}, fmt.Errorf("unexpected length from exchange rate query: %d", len(queryRes))
	}

	currencyExchange := models.Exchange{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         queryRes[0].Rate,
		Date:         time,
	}

	return currencyExchange, nil
}

func GetExchangeRateDBEntry(fromCurrency, toCurrency string, time time.Time) (models.ExchangeDBEntry, error) {
	exchangeData, err := GetExchangeRate(fromCurrency, toCurrency, time)
	if err != nil {
		return models.ExchangeDBEntry{}, err
	}

	return models.ExchangeDBEntry{
		FromCurrency: exchangeData.FromCurrency,
		ToCurrency:   exchangeData.ToCurrency,
		Rate:         exchangeData.Rate,
		Date:         exchangeData.Date,
	}, nil
}
