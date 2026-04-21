package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func PaymentResponseFromDB(payment dbmodels.PaymentDBEntry) apimodels.PaymentResponse {
	return apimodels.PaymentResponse{
		ID:          payment.ID,
		Amount:      payment.Amount,
		Description: payment.Description,
		Date:        payment.Date,
		Payer:       UserSummaryFromDB(payment.Payer),
		Debtors:     UserSummariesFromDB(payment.Debtors),
		Exchange:    ExchangeSummaryFromDB(payment.Exchange),
	}
}

func PaymentResponsesFromDB(payments []dbmodels.PaymentDBEntry) []apimodels.PaymentResponse {
	responses := make([]apimodels.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, PaymentResponseFromDB(payment))
	}

	return responses
}
