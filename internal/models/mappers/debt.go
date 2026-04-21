package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func DebtResponseFromDB(debt dbmodels.DebtDBEntry) apimodels.DebtResponse {
	return apimodels.DebtResponse{
		ID:           debt.ID,
		OwedByUserID: debt.OwedByUserId,
		OwedToUserID: debt.OwedToUserId,
		Amount:       debt.Amount,
		Currency:     debt.Currency,
	}
}

func DebtResponsesFromDB(debts []dbmodels.DebtDBEntry) []apimodels.DebtResponse {
	responses := make([]apimodels.DebtResponse, 0, len(debts))
	for _, debt := range debts {
		responses = append(responses, DebtResponseFromDB(debt))
	}

	return responses
}
