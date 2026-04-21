package mappers

import (
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
)

func UserSummaryFromDB(user dbmodels.UserDBEntry) apimodels.UserSummary {
	return apimodels.UserSummary{
		ID:   user.ID,
		Name: user.Name,
	}
}

func UserSummariesFromDB(users []dbmodels.UserDBEntry) []apimodels.UserSummary {
	summaries := make([]apimodels.UserSummary, 0, len(users))
	for _, user := range users {
		summaries = append(summaries, UserSummaryFromDB(user))
	}

	return summaries
}
