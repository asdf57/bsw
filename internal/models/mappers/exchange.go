package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func ExchangeSummaryFromDB(exchange dbmodels.ExchangeDBEntry) apimodels.ExchangeSummary {
	return apimodels.ExchangeSummary{
		ID:           exchange.ID,
		FromCurrency: exchange.FromCurrency,
		ToCurrency:   exchange.ToCurrency,
		Date:         exchange.Date,
		Rate:         exchange.Rate,
	}
}
