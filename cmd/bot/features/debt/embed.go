package debt

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

func CreateSettleUpEmbed(owedBy string, owedTo string, settledCount int64, settlements []apimodels.SettlementResponse, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Settle Up Failed", Description: fmt.Sprintf("An error occurred while settling up for user:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	fields := []*discordgo.MessageEmbedField{
		{Name: "Owed By", Value: "`" + owedBy + "`", Inline: true},
		{Name: "Settled Debts", Value: fmt.Sprintf("`%d`", settledCount), Inline: true},
	}
	if strings.TrimSpace(owedTo) != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Owed To", Value: "`" + owedTo + "`", Inline: true})
	}
	if len(settlements) > 0 {
		lines := make([]string, 0, len(settlements))
		for _, settlement := range settlements {
			lines = append(lines, fmt.Sprintf("%s -> %s: %s",
				settlement.OwedByUser.Name,
				settlement.OwedToUser.Name,
				shared.FormatAmount(apimodels.Payment{Amount: settlement.Amount.InexactFloat64(), Currency: settlement.Currency}),
			))
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Recorded Settlements", Value: strings.Join(lines, "\n"), Inline: false})
	}
	return &discordgo.MessageEmbed{Title: "Settled Up", Description: "You're all settled up!", Color: 0x2ECC71, Fields: fields, Timestamp: time.Now().Format(time.RFC3339)}
}

func ListMessage(debts []apimodels.DebtResponse) string {
	const maxRows = 20

	if len(debts) == 0 {
		return "**Debts**\nAll settled up."
	}

	lines := []string{"**Debts**"}
	limit := min(maxRows, len(debts))

	users := []string{}
	seenUsers := map[string]struct{}{}

	userDebts := make(map[string][]apimodels.DebtResponse)
	for _, d := range debts {
		userDebts[d.OwedByUser.Name] = append(userDebts[d.OwedByUser.Name], d)
		if _, ok := seenUsers[d.OwedByUser.Name]; !ok {
			users = append(users, d.OwedByUser.Name)
			seenUsers[d.OwedByUser.Name] = struct{}{}
		}
	}

	// Sort users lexicographically
	sort.Strings(users)

	numDebtsDisplayed := 0

	for _, user := range users {
		debts := userDebts[user]
		for _, d := range debts {
			if numDebtsDisplayed >= limit {
				break
			}
			lines = append(lines,
				fmt.Sprintf("• **%s -> %s:** `%s`",
					d.OwedByUser.Name,
					d.OwedToUser.Name,
					shared.FormatAmount(apimodels.Payment{
						Amount:   d.Amount.InexactFloat64(),
						Currency: d.Currency,
					},
					),
				))

			numDebtsDisplayed++
		}
	}

	if len(debts) > maxRows {
		lines = append(lines, "", fmt.Sprintf("_Showing first %d of %d debts._", maxRows, len(debts)))
	}
	return strings.Join(lines, "\n")
}

func CreateStatsEmbed(user string, requestedCurrency string, stats *apimodels.UserStatsResponse, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Stats Failed", Description: fmt.Sprintf("An error occurred while fetching stats:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	if stats == nil {
		return &discordgo.MessageEmbed{Title: "Stats Failed", Description: "No stats response was returned.", Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}

	currency := stats.Currency
	if strings.TrimSpace(currency) == "" {
		currency = requestedCurrency
	}
	displayUser := stats.User.Name
	if strings.TrimSpace(displayUser) == "" {
		displayUser = user
	}

	totalSpent := shared.FormatAmount(apimodels.Payment{Amount: stats.TotalSpent.InexactFloat64(), Currency: currency})
	spentOut := shared.FormatAmount(apimodels.Payment{Amount: stats.SpentOut.InexactFloat64(), Currency: currency})

	return &discordgo.MessageEmbed{
		Title:       "Spending Stats",
		Description: "Generated in " + strings.ToUpper(currency) + ".",
		Color:       0x3B82F6,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: "`" + displayUser + "`", Inline: true},
			{Name: "Total Spent", Value: "`" + totalSpent + "`", Inline: true},
			{Name: "Spent Out", Value: "`" + spentOut + "`", Inline: true},
			{Name: "Payments Included", Value: fmt.Sprintf("`%d`", stats.PaymentsIncluded), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func ListSettlementsMessage(settlements []apimodels.SettlementResponse) string {
	const maxRows = 20
	if len(settlements) == 0 {
		return "**Settlements**\nNo settlements recorded."
	}

	lines := []string{"**Settlements**"}
	limit := min(maxRows, len(settlements))
	for _, settlement := range settlements[:limit] {
		status := ""
		if settlement.ReversedAt != nil {
			status = " (reversed)"
		}
		lines = append(lines, fmt.Sprintf("• **%s -> %s:** `%s` on `%s`",
			settlement.OwedByUser.Name,
			settlement.OwedToUser.Name,
			shared.FormatAmount(apimodels.Payment{Amount: settlement.Amount.InexactFloat64(), Currency: settlement.Currency}),
			settlement.Date.Local().Format("2006-01-02")+status,
		))
	}
	if len(settlements) > maxRows {
		lines = append(lines, "", fmt.Sprintf("_Showing first %d of %d settlements._", maxRows, len(settlements)))
	}
	return strings.Join(lines, "\n")
}
