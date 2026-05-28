package debt

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

const (
	SettleModalID       = "settle_modal"
	SettleOwedByID      = "settle_owedby"
	SettleOwedToID      = "settle_owedto"
	StatsModalID        = "stats_modal"
	StatsUserID         = "stats_user"
	StatsCurrencyID     = "stats_currency"
	defaultStatCurrency = "USD"
)

func OpenSettleModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	optional := false
	users, err := shared.GetUsers()
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch users failed: "+err.Error())
	}
	userOpts := userOptions(users)
	if len(userOpts) == 0 {
		return shared.RespondWithMessage(s, i, "No users available. Add users first with `/adduser`.")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: SettleModalID,
			Title:    "Settle Up",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Owed By", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: SettleOwedByID, Placeholder: "Select user who paid", MaxValues: 1, Required: &required, Options: userOpts}},
				discordgo.Label{Label: "Owed To", Description: "Optional. Leave empty to settle all debts this user owes.", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: SettleOwedToID, Placeholder: "Select one user", MaxValues: 1, Required: &optional, Options: userOpts}},
			},
		},
	})
}

func HandleSettleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	owedBy := strings.TrimSpace(values[SettleOwedByID])
	owedTo := strings.TrimSpace(values[SettleOwedToID])
	if owedBy == "" {
		return shared.RespondWithMessage(s, i, "Select who is settling up.")
	}
	return settle(s, i, owedBy, owedTo)
}

func settle(s *discordgo.Session, i *discordgo.InteractionCreate, owedBy string, owedTo string) error {
	settleResp, err := shared.SettleDebts(owedBy, owedTo)
	settledCount := int64(0)
	var settlements []apimodels.SettlementResponse
	if settleResp != nil {
		settledCount = settleResp.SettledCount
		settlements = settleResp.Settlements
	}
	embed := CreateSettleUpEmbed(owedBy, owedTo, settledCount, settlements, err)
	if err := shared.RespondWithEmbed(s, i, embed); err != nil {
		return err
	}
	if err != nil {
		return nil
	}
	s.ChannelMessageSend(i.ChannelID, "So proud of you!")
	switch rand.IntN(6) {
	case 0:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/iYkARYAbfTg.gif")
	case 1:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/view/frieren-pat-frieren-pat-pat-frieren-pat-pat-gif-16450085803075970062")
	case 2:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/view/anime-head-pat-anime-gif-6292920416547557855")
	case 3:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/view/frieren-pats-heither's-head-sousou-no-frieren-headpat-frieren-heither-gif-9445174606794519653")
	case 4:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/view/frieren-headpat-sein-rooftop-praise-sousou-no-frieren-patpatpat-gif-18014250646250877804")
	case 5:
		s.ChannelMessageSend(i.ChannelID, "https://tenor.com/view/frieren-crying-frieren-crying-gif-199542684162590099")
	}
	return nil
}

func HandleGetDebts(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	currency := "USD"
	for _, opt := range data.Options {
		if opt.Name == "currency" {
			currency = opt.StringValue()
			break
		}
	}
	debts, err := shared.GetDebtResponses(currency)
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch debts failed: "+err.Error())
	}
	return shared.RespondWithMessage(s, i, ListMessage(debts))
}

func OpenStatsModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	optional := false
	users, err := shared.GetUsers()
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch users failed: "+err.Error())
	}
	userOpts := userOptions(users)
	if len(userOpts) == 0 {
		return shared.RespondWithMessage(s, i, "No users available. Add users first with `/adduser`.")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: StatsModalID,
			Title:    "Spending Stats",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "User", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: StatsUserID, Placeholder: "Select user", MaxValues: 1, Required: &required, Options: userOpts}},
				discordgo.Label{Label: "Currency", Description: "Optional. Defaults to USD.", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: StatsCurrencyID, Placeholder: "Select currency", MaxValues: 1, Required: &optional, Options: currencyOptions()}},
			},
		},
	})
}

func HandleStatsModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	user := strings.TrimSpace(values[StatsUserID])
	currency := strings.TrimSpace(values[StatsCurrencyID])
	if currency == "" {
		currency = defaultStatCurrency
	}
	if user == "" {
		return shared.RespondWithMessage(s, i, "Select a user.")
	}

	stats, err := shared.GetUserStats(user, currency)
	return shared.RespondWithEmbed(s, i, CreateStatsEmbed(user, currency, stats, err))
}

func HandleSettlements(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	settlements, err := shared.GetSettlements()
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch settlements failed: "+err.Error())
	}
	return shared.RespondWithMessage(s, i, ListSettlementsMessage(settlements))
}

func HandleReverseSettlement(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var settlementID int64
	for _, opt := range data.Options {
		if opt.Name == "id" {
			settlementID = opt.IntValue()
			break
		}
	}
	if settlementID <= 0 {
		return shared.RespondWithMessage(s, i, "Provide a valid settlement ID.")
	}

	settlement, err := shared.ReverseSettlement(uint(settlementID))
	if err != nil {
		return shared.RespondWithMessage(s, i, "reverse settlement failed: "+err.Error())
	}

	return shared.RespondWithMessage(s, i, fmt.Sprintf("Reversed settlement #%d: %s -> %s for `%s`.",
		settlement.ID,
		settlement.OwedByUser.Name,
		settlement.OwedToUser.Name,
		shared.FormatAmount(apimodels.Payment{Amount: settlement.Amount.InexactFloat64(), Currency: settlement.Currency}),
	))
}

func userOptions(users []apimodels.UserSummary) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, len(users))
	for _, user := range users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			continue
		}
		option := discordgo.SelectMenuOption{Label: name, Value: name}
		if strings.TrimSpace(user.DiscordHandle) != "" {
			option.Description = user.DiscordHandle
		}
		options = append(options, option)
		if len(options) == 25 {
			break
		}
	}
	return options
}

func currencyOptions() []discordgo.SelectMenuOption {
	return []discordgo.SelectMenuOption{
		{Label: "USD", Value: "USD", Default: true},
		{Label: "EUR", Value: "EUR"},
		{Label: "GBP", Value: "GBP"},
		{Label: "CAD", Value: "CAD"},
		{Label: "JPY", Value: "JPY"},
	}
}
