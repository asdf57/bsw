package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func SettlementResponseFromDB(settlement dbmodels.SettlementDBEntry) apimodels.SettlementResponse {
	return apimodels.SettlementResponse{
		ID:         settlement.ID,
		OwedByUser: UserSummaryFromDB(settlement.OwedBy),
		OwedToUser: UserSummaryFromDB(settlement.OwedTo),
		Amount:     settlement.Amount,
		Currency:   settlement.Currency,
		Date:       settlement.Date,
		ReversedAt: settlement.ReversedAt,
	}
}

func SettlementResponsesFromDB(settlements []dbmodels.SettlementDBEntry) []apimodels.SettlementResponse {
	responses := make([]apimodels.SettlementResponse, 0, len(settlements))
	for _, settlement := range settlements {
		responses = append(responses, SettlementResponseFromDB(settlement))
	}
	return responses
}
