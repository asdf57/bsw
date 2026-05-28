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
		Currency:    payment.Currency,
		Tags:        TagsFromDB(payment.Tags),
	}
}

func PaymentResponsesFromDB(payments []dbmodels.PaymentDBEntry) []apimodels.PaymentResponse {
	responses := make([]apimodels.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, PaymentResponseFromDB(payment))
	}

	return responses
}

func TagsFromDB(tags []dbmodels.TagDBEntry) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		out = append(out, tag.Name)
	}
	return out
}
