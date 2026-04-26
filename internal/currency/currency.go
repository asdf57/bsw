package currency

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func NormalizeExchangeDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func GetExchangeRate(fromCurrency, toCurrency string, time time.Time) (apimodels.Exchange, error) {
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
