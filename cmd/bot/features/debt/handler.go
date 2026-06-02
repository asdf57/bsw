package debt

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

const (
	SettleAllModalID           = "settleall_modal"
	SettleCustomOwedBySelectID = "settlecustom_select_owedby"
	SettleCustomOwedToPrefix   = "settlecustom_select_owedto:"
	SettleAmountPrefix         = "settlecustom_amount_modal:"
	SettleOwedByID             = "settle_owedby"
	SettleOwedToID             = "settle_owedto"
	SettleAmountID             = "settle_amount"
	StatsModalID               = "stats_modal"
	StatsUserID                = "stats_user"
	StatsCurrencyID            = "stats_currency"
	defaultStatCurrency        = "USD"
)

func settleAmountModalID(owedBy string, owedTo string, maxAmount string, currency string) string {
	values := url.Values{}
	values.Set("owedBy", strings.TrimSpace(owedBy))
	values.Set("owedTo", strings.TrimSpace(owedTo))
	values.Set("max", strings.TrimSpace(maxAmount))
	values.Set("currency", strings.TrimSpace(currency))
	return SettleAmountPrefix + values.Encode()
}

func settleAmountContext(customID string) (string, string, string, string, bool) {
	if !strings.HasPrefix(customID, SettleAmountPrefix) {
		return "", "", "", "", false
	}
	values, err := url.ParseQuery(strings.TrimPrefix(customID, SettleAmountPrefix))
	if err != nil {
		return "", "", "", "", false
	}
	owedBy := strings.TrimSpace(values.Get("owedBy"))
	owedTo := strings.TrimSpace(values.Get("owedTo"))
	maxAmount := strings.TrimSpace(values.Get("max"))
	currency := strings.TrimSpace(values.Get("currency"))
	return owedBy, owedTo, maxAmount, currency, owedBy != "" && owedTo != "" && maxAmount != ""
}

func settleCustomOwedToSelectID(owedBy string) string {
	return SettleCustomOwedToPrefix + url.QueryEscape(strings.TrimSpace(owedBy))
}

func owedByFromSettleCustomOwedToSelectID(customID string) (string, bool) {
	if !strings.HasPrefix(customID, SettleCustomOwedToPrefix) {
		return "", false
	}
	owedBy, err := url.QueryUnescape(strings.TrimPrefix(customID, SettleCustomOwedToPrefix))
	if err != nil {
		return "", false
	}
	owedBy = strings.TrimSpace(owedBy)
	return owedBy, owedBy != ""
}

func OpenSettleAllModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
			CustomID: SettleAllModalID,
			Title:    "Settle All",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Owed By", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: SettleOwedByID, Placeholder: "Select user who paid", MaxValues: 1, Required: &required, Options: userOpts}},
				discordgo.Label{Label: "Owed To", Description: "Optional. Leave empty to settle all debts this user owes.", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: SettleOwedToID, Placeholder: "Select one user", MaxValues: 1, Required: &optional, Options: userOpts}},
			},
		},
	})
}

func HandleSettleAllModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	owedBy := strings.TrimSpace(values[SettleOwedByID])
	owedTo := strings.TrimSpace(values[SettleOwedToID])
	if owedBy == "" {
		return shared.RespondWithMessage(s, i, "Select who is settling up.")
	}
	return settle(s, i, owedBy, owedTo)
}

func OpenSettleCustomStart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	users, err := shared.GetUsers()
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch users failed: "+err.Error())
	}
	userOpts := userOptions(users)
	if len(userOpts) == 0 {
		return shared.RespondWithMessage(s, i, "No users available. Add users first with `/adduser`.")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select who is settling up.",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: SettleCustomOwedBySelectID, Placeholder: "Select user", MaxValues: 1, Required: &required, Options: userOpts}}},
			},
		},
	})
}

func HandleSettleCustomOwedBySelected(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	if data.CustomID != SettleCustomOwedBySelectID {
		return nil
	}
	if len(data.Values) == 0 {
		return shared.RespondWithMessage(s, i, "Select who is settling up.")
	}
	owedBy := strings.TrimSpace(data.Values[0])
	if owedBy == "" {
		return shared.RespondWithMessage(s, i, "Select who is settling up.")
	}
	users, err := shared.GetUsers()
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch users failed: "+err.Error())
	}
	options := userOptions(users)
	required := true
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Select who `%s` is settling with.", owedBy),
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: settleCustomOwedToSelectID(owedBy), Placeholder: "Select user", MaxValues: 1, Required: &required, Options: options}}},
			},
		},
	})
}

func HandleSettleCustomOwedToSelected(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	owedBy, ok := owedByFromSettleCustomOwedToSelectID(data.CustomID)
	if !ok {
		return nil
	}
	if len(data.Values) == 0 {
		return shared.RespondWithMessage(s, i, "Select who to settle with.")
	}
	owedTo := strings.TrimSpace(data.Values[0])
	if owedTo == "" || strings.EqualFold(owedBy, owedTo) {
		return shared.RespondWithMessage(s, i, "Select a different user to settle with.")
	}
	return openSettleAmountModal(s, i, owedBy, owedTo)
}

func openSettleAmountModal(s *discordgo.Session, i *discordgo.InteractionCreate, owedBy string, owedTo string) error {
	debt, err := debtBetweenUsers(owedBy, owedTo, "USD")
	if err != nil {
		return shared.RespondWithMessage(s, i, err.Error())
	}
	if debt == nil {
		return shared.RespondWithMessage(s, i, fmt.Sprintf("%s does not owe %s anything.", owedBy, owedTo))
	}

	required := true
	amount := debt.Amount.StringFixed(2)
	displayAmount := shared.FormatAmount(apimodels.Payment{Amount: debt.Amount.InexactFloat64(), Currency: debt.Currency})
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: settleAmountModalID(owedBy, owedTo, amount, debt.Currency),
			Title:    settleAmountModalTitle(owedBy, owedTo),
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Settlement Amount", Description: fmt.Sprintf("%s owes %s %s. Min 0.01, max %s.", owedBy, owedTo, displayAmount, displayAmount), Component: discordgo.TextInput{CustomID: SettleAmountID, Style: discordgo.TextInputShort, Value: amount, Required: &required}},
			},
		},
	})
}

func settleAmountModalTitle(owedBy string, owedTo string) string {
	title := fmt.Sprintf("Settling %s and %s", owedBy, owedTo)
	if len(title) <= 45 {
		return title
	}
	return "Settle Debt"
}

func HandleSettleAmountModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	owedBy, owedTo, maxAmount, currency, ok := settleAmountContext(i.ModalSubmitData().CustomID)
	if !ok {
		return shared.RespondWithMessage(s, i, "invalid settlement amount metadata; please run `/settle` again")
	}
	values := shared.ExtractModalValues(i.ModalSubmitData())
	amount := strings.TrimSpace(values[SettleAmountID])
	if err := validateSettlementAmount(amount, maxAmount); err != nil {
		return shared.RespondWithMessage(s, i, err.Error())
	}
	_ = currency
	return settleAmount(s, i, owedBy, owedTo, amount)
}

func debtBetweenUsers(owedBy string, owedTo string, currency string) (*apimodels.DebtResponse, error) {
	debts, err := shared.GetDebtResponses(currency)
	if err != nil {
		return nil, fmt.Errorf("fetch debts failed: %w", err)
	}
	for _, debt := range debts {
		if strings.EqualFold(debt.OwedByUser.Name, owedBy) && strings.EqualFold(debt.OwedToUser.Name, owedTo) {
			return &debt, nil
		}
	}
	return nil, nil
}

func validateSettlementAmount(amount string, maxAmount string) error {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(amount), 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %s", err.Error())
	}
	max, err := strconv.ParseFloat(strings.TrimSpace(maxAmount), 64)
	if err != nil {
		return fmt.Errorf("invalid max settlement amount")
	}
	if parsed < 0.01 {
		return fmt.Errorf("settlement amount must be at least 0.01")
	}
	if parsed > max {
		return fmt.Errorf("settlement amount cannot exceed %.2f", max)
	}
	return nil
}

func settle(s *discordgo.Session, i *discordgo.InteractionCreate, owedBy string, owedTo string) error {
	settleResp, err := shared.SettleDebts(owedBy, owedTo)
	return respondToSettlement(s, i, owedBy, owedTo, settleResp, err)
}

func settleAmount(s *discordgo.Session, i *discordgo.InteractionCreate, owedBy string, owedTo string, amount string) error {
	settleResp, err := shared.SettleDebtAmount(owedBy, owedTo, amount)
	return respondToSettlement(s, i, owedBy, owedTo, settleResp, err)
}

func respondToSettlement(s *discordgo.Session, i *discordgo.InteractionCreate, owedBy string, owedTo string, settleResp *apimodels.SettleDebtsResponse, err error) error {
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
