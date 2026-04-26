package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func DebtResponseFromDB(debt dbmodels.DebtDBEntry, usersByID map[uint]dbmodels.UserDBEntry) apimodels.DebtResponse {
	owedById := debt.UserLowId
	owedToId := debt.UserHighId

	if debt.NetAmount.IsNegative() {
		owedById = debt.UserHighId
		owedToId = debt.UserLowId
	}

	return apimodels.DebtResponse{
		ID:         debt.ID,
		OwedByUser: UserSummaryFromDB(usersByID[owedById]),
		OwedToUser: UserSummaryFromDB(usersByID[owedToId]),
		Amount:     debt.NetAmount.Abs(),
		Currency:   debt.Currency,
	}
}

func DebtResponsesFromDB(debts []dbmodels.DebtDBEntry, usersByID map[uint]dbmodels.UserDBEntry) []apimodels.DebtResponse {
	responses := make([]apimodels.DebtResponse, 0, len(debts))
	for _, debt := range debts {
		responses = append(responses, DebtResponseFromDB(debt, usersByID))
	}

	return responses
}
