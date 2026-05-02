package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

const (
	defaultPaymentCurrency = "USD"
)

func main() {
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		log.Fatal("missing required env var: DISCORD_BOT_TOKEN")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("failed to create discord session: %v", err)
	}
	defer dg.Close()

	dg.AddHandler(onMessage)
	dg.AddHandler(onInteractionCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	if err := dg.Open(); err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}

	if err := registerCommands(dg); err != nil {
		log.Fatalf("failed to register commands: %v", err)
	}

	log.Println("discord bot is connected; waiting for shutdown signal")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received; closing discord session")
}

func getPayments() (string, error) {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return "", fmt.Errorf("missing required env var: API_URL")
	}

	resp, err := http.Get(apiURL + "/api/v1/payment/all")
	if err != nil {
		return "", fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body response: %w", err)
	}

	return string(body), nil
}

func getPaymentResponses() ([]apimodels.PaymentResponse, error) {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return nil, fmt.Errorf("missing required env var: API_URL")
	}

	resp, err := http.Get(apiURL + "/api/v1/payment/all")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var payments []apimodels.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payments); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	return payments, nil
}

func getUsers() ([]apimodels.UserSummary, error) {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return nil, fmt.Errorf("missing required env var: API_URL")
	}

	resp, err := http.Get(apiURL + "/api/v1/user")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var users []apimodels.UserSummary
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	return users, nil
}

func getDebtResponses() ([]apimodels.DebtResponse, error) {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return nil, fmt.Errorf("missing required env var: API_URL")
	}

	resp, err := http.Get(apiURL + "/api/v1/debts")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var debts []apimodels.DebtResponse
	if err := json.NewDecoder(resp.Body).Decode(&debts); err != nil {
		return nil, fmt.Errorf("error decoding debt response: %w", err)
	}

	return debts, nil
}

func deletePayment(paymentId uint) error {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return fmt.Errorf("missing required env var: API_URL")
	}

	url := fmt.Sprintf("%s/api/v1/payment/%d", apiURL, paymentId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing payment deletion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func createPayment(payment *apimodels.Payment) error {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return fmt.Errorf("missing required env var: API_URL")
	}

	body, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshall payment json")
	}

	resp, err := http.Post(apiURL+"/api/v1/payment", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create payment")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func registerCommands(s *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "payment",
			Description: "Create a new payment",
		},
		{
			Name:        "getpayments",
			Description: "Show all recorded payments",
		},
		{
			Name:        "getdebts",
			Description: "Show all current debts",
		},
		{
			Name:        "delpayment",
			Description: "Delete a payment",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Payment ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "adduser",
			Description: "Create a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "User's name",
					Required:    true,
				},
			},
		},
		{
			Name:        "settle",
			Description: "Settle up",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "User who is settling up",
					Required:    true,
				},
			},
		},
	}

	for _, cmd := range commands {
		if _, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd); err != nil {
			return err
		}
	}

	return nil
}

func normalizePaymentCurrency(currency string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(currency))
	if cleaned == "" {
		cleaned = defaultPaymentCurrency
	}

	return cleaned
}

func openAddPaymentModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	minValues := 1

	// Get all users
	users, err := getUsers()
	if err != nil {
		return fmt.Errorf("Failed to fetch users!")
	}

	// Extract users
	allUsers := make([]discordgo.SelectMenuOption, 0, len(users))
	for _, u := range users {
		name := strings.TrimSpace(u.Name)
		if name == "" {
			continue
		}
		allUsers = append(allUsers, discordgo.SelectMenuOption{Label: name, Value: name})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "addpayment_modal",
			Title:    "Add Payment",
			Components: []discordgo.MessageComponent{
				discordgo.Label{
					Label: "Amount",
					Component: discordgo.TextInput{
						CustomID:    "amount",
						Style:       discordgo.TextInputShort,
						Placeholder: "10.00",
						Required:    &required,
					},
				},
				discordgo.Label{
					Label: "Payer",
					Component: discordgo.SelectMenu{
						CustomID:    "payer",
						Placeholder: "Select payer",
						MinValues:   &minValues,
						MaxValues:   1,
						Required:    &required,
						Options:     allUsers,
					},
				},
				discordgo.Label{
					Label: "Description",
					Component: discordgo.TextInput{
						CustomID:    "description",
						Style:       discordgo.TextInputShort,
						Placeholder: "Dinner",
						Required:    &required,
					},
				},
				discordgo.Label{
					Label:       "Currency",
					Description: "Choose the payment currency.",
					Component: discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						CustomID:    "currency",
						Placeholder: "Select a currency",
						MaxValues:   1,
						Required:    &required,
						Options: []discordgo.SelectMenuOption{
							{Label: "USD", Value: "USD", Default: true},
							{Label: "EUR", Value: "EUR"},
							{Label: "GBP", Value: "GBP"},
							{Label: "CAD", Value: "CAD"},
							{Label: "JPY", Value: "JPY"},
						},
					},
				},
				discordgo.Label{
					Label: "Debtors",
					Component: discordgo.SelectMenu{
						CustomID:    "debtors",
						Placeholder: "Select debtors",
						MinValues:   &minValues,
						MaxValues:   len(allUsers),
						Required:    &required,
						Options:     allUsers,
					},
				},
			},
		},
	})
}

func extractModalValues(data discordgo.ModalSubmitInteractionData) map[string]string {
	values := make(map[string]string)

	for _, comp := range data.Components {
		switch component := comp.(type) {
		case *discordgo.Label:
			switch child := component.Component.(type) {
			case *discordgo.TextInput:
				values[child.CustomID] = child.Value
			case *discordgo.SelectMenu:
				if len(child.Values) > 0 {
					values[child.CustomID] = strings.Join(child.Values, ",")
				}
			}
		case *discordgo.ActionsRow:
			for _, rowComp := range component.Components {
				if input, ok := rowComp.(*discordgo.TextInput); ok {
					values[input.CustomID] = input.Value
				}
				if selectMenu, ok := rowComp.(*discordgo.SelectMenu); ok && len(selectMenu.Values) > 0 {
					values[selectMenu.CustomID] = strings.Join(selectMenu.Values, ",")
				}
			}
		}
	}

	return values
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}

func respondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

func respondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func stringPtr(value string) *string {
	return &value
}

func respondDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func formatPaymentAmount(payment apimodels.Payment) string {
	currency := strings.ToUpper(strings.TrimSpace(payment.ToExchangeRate))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(payment.FromExchangeRate))
	}
	if currency == "" {
		return fmt.Sprintf("%.2f", payment.Amount)
	}

	return fmt.Sprintf("%.2f %s", payment.Amount, currency)
}

func formatDebtors(debtors []string) string {
	if len(debtors) == 0 {
		return ""
	}

	lines := make([]string, 0, len(debtors))
	for _, debtor := range debtors {
		lines = append(lines, "• "+debtor)
	}

	return strings.Join(lines, "\n")
}

func paymentCreatedEmbed(payment apimodels.Payment, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{
			Title:       "Payment Creation Failed",
			Description: fmt.Sprintf("An error occurred while recording the payment:\n%s", err.Error()),
			Color:       0xE74C3C,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}

	description := strings.TrimSpace(payment.Description)
	if description == "" {
		description = "_none_"
	}

	return &discordgo.MessageEmbed{
		Title:       "Payment Created",
		Description: "The payment was recorded successfully.",
		Color:       0x2ECC71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Amount",
				Value:  "`" + formatPaymentAmount(payment) + "`",
				Inline: true,
			},
			{
				Name:   "Payer",
				Value:  payment.Payer,
				Inline: true,
			},
			{
				Name:   "Debtors",
				Value:  formatDebtors(payment.Debtors),
				Inline: false,
			},
			{
				Name:   "Description",
				Value:  description,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func formatPaymentResponseAmount(payment apimodels.PaymentResponse) string {
	fromCurrency := strings.ToUpper(strings.TrimSpace(payment.Exchange.FromCurrency))
	toCurrency := strings.ToUpper(strings.TrimSpace(payment.Exchange.ToCurrency))
	amount := payment.Amount.StringFixed(2)

	if toCurrency == "" && fromCurrency == "" {
		return amount
	}

	if toCurrency == "" || toCurrency == fromCurrency {
		return fmt.Sprintf("%s %s", amount, fromCurrency)
	}

	return fmt.Sprintf("%s %s -> %s", amount, fromCurrency, toCurrency)
}

func formatPaymentResponseDebtors(debtors []apimodels.UserSummary) string {
	if len(debtors) == 0 {
		return ""
	}

	names := make([]string, 0, len(debtors))
	for _, debtor := range debtors {
		names = append(names, debtor.Name)
	}

	return strings.Join(names, ", ")
}

func paymentFieldValue(payment apimodels.PaymentResponse) string {
	lines := []string{
		fmt.Sprintf("**Payer:** %s", payment.Payer.Name),
		fmt.Sprintf("**Debtors:** %s", formatPaymentResponseDebtors(payment.Debtors)),
		fmt.Sprintf("**Date:** %s", payment.Date.Local().Format("2006-01-02 15:04")),
	}

	description := strings.TrimSpace(payment.Description)
	if description != "" {
		lines = append(lines, fmt.Sprintf("**Description:** %s", description))
	}

	if payment.Exchange.FromCurrency != "" && payment.Exchange.ToCurrency != "" && payment.Exchange.FromCurrency != payment.Exchange.ToCurrency {
		lines = append(lines, fmt.Sprintf("**Rate:** 1 %s = %.6f %s", payment.Exchange.FromCurrency, payment.Exchange.Rate, payment.Exchange.ToCurrency))
	}

	return strings.Join(lines, "\n")
}

func paymentThreadEmbed(payment apimodels.PaymentResponse) *discordgo.MessageEmbed {
	description := strings.TrimSpace(payment.Description)
	if description == "" {
		description = "_none_"
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Payer",
			Value:  payment.Payer.Name,
			Inline: true,
		},
		{
			Name:   "Date",
			Value:  payment.Date.Local().Format("2006-01-02 15:04"),
			Inline: true,
		},
		{
			Name:   "Debtors",
			Value:  formatPaymentResponseDebtors(payment.Debtors),
			Inline: false,
		},
		{
			Name:   "Description",
			Value:  description,
			Inline: false,
		},
	}

	if payment.Exchange.FromCurrency != "" && payment.Exchange.ToCurrency != "" && payment.Exchange.FromCurrency != payment.Exchange.ToCurrency {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Exchange Rate",
			Value:  fmt.Sprintf("1 %s = %.6f %s", payment.Exchange.FromCurrency, payment.Exchange.Rate, payment.Exchange.ToCurrency),
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("#%d • %s", payment.ID, formatPaymentResponseAmount(payment)),
		Color:     0x3B82F6,
		Fields:    fields,
		Timestamp: payment.Date.Format(time.RFC3339),
	}
}

func createPaymentsThread(s *discordgo.Session, channelID string, payments []apimodels.PaymentResponse) (*discordgo.Channel, error) {
	thread, err := s.ThreadStart(channelID, "Payments", discordgo.ChannelTypeGuildPublicThread, 1440)
	if err != nil {
		return nil, fmt.Errorf("start payments thread: %w", err)
	}

	messages, err := s.ChannelMessages(channelID, 10, "", "", "")
	if err != nil {
		return nil, err
	}

	for _, msg := range messages {
		if msg.Type == discordgo.MessageTypeThreadCreated {
			if err := s.ChannelMessageDelete(channelID, msg.ID); err != nil {
				return nil, err
			}
			continue
		}

		if msg.Author != nil &&
			msg.Author.ID == s.State.User.ID &&
			strings.HasPrefix(msg.Content, "Posted ") &&
			strings.Contains(msg.Content, " payments in ") {
			if err := s.ChannelMessageDelete(channelID, msg.ID); err != nil {
				return nil, err
			}
		}
	}

	if len(payments) == 0 {
		if _, err := s.ChannelMessageSend(thread.ID, "No payments have been recorded yet."); err != nil {
			return nil, fmt.Errorf("post empty payments message: %w", err)
		}

		return thread, nil
	}

	for _, payment := range payments {
		if _, err := s.ChannelMessageSendEmbed(thread.ID, paymentThreadEmbed(payment)); err != nil {
			return nil, fmt.Errorf("post payment %d to thread: %w", payment.ID, err)
		}
	}

	return thread, nil
}

func truncateCell(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func padCell(value string, width int) string {
	return fmt.Sprintf("%-*s", width, truncateCell(value, width))
}

func paymentListTable(payments []apimodels.PaymentResponse) string {
	if len(payments) == 0 {
		return "No payments have been recorded yet."
	}

	const (
		idWidth          = 4
		payerWidth       = 12
		debtorsWidth     = 22
		amountWidth      = 16
		dateWidth        = 16
		descriptionWidth = 24
		maxRows          = 12
	)

	lines := []string{
		fmt.Sprintf(
			"%s  %s  %s  %s  %s  %s",
			padCell("ID", idWidth),
			padCell("PAYER", payerWidth),
			padCell("DEBTORS", debtorsWidth),
			padCell("AMOUNT", amountWidth),
			padCell("DATE", dateWidth),
			padCell("DESCRIPTION", descriptionWidth),
		),
		fmt.Sprintf(
			"%s  %s  %s  %s  %s  %s",
			strings.Repeat("-", idWidth),
			strings.Repeat("-", payerWidth),
			strings.Repeat("-", debtorsWidth),
			strings.Repeat("-", amountWidth),
			strings.Repeat("-", dateWidth),
			strings.Repeat("-", descriptionWidth),
		),
	}

	limit := min(len(payments), maxRows)

	for _, payment := range payments[:limit] {
		dateValue := payment.Date.Local().Format("2006-01-02 15:04")
		description := payment.Description
		if strings.TrimSpace(description) == "" {
			description = "-"
		}

		lines = append(lines, fmt.Sprintf(
			"%s  %s  %s  %s  %s  %s",
			padCell(strconv.FormatUint(uint64(payment.ID), 10), idWidth),
			padCell(payment.Payer.Name, payerWidth),
			padCell(formatPaymentResponseDebtors(payment.Debtors), debtorsWidth),
			padCell(formatPaymentResponseAmount(payment), amountWidth),
			padCell(dateValue, dateWidth),
			padCell(description, descriptionWidth),
		))
	}

	if len(payments) > maxRows {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Showing first %d of %d payments.", maxRows, len(payments)))
	}

	return "```text\n" + strings.Join(lines, "\n") + "\n```"
}

func debtListMessage(debts []apimodels.DebtResponse) string {
	if len(debts) == 0 {
		return "**Debts**\nAll settled up."
	}

	const maxRows = 20

	lines := []string{"**Debts**"}
	limit := min(len(debts), maxRows)

	for _, debt := range debts[:limit] {
		lines = append(lines, fmt.Sprintf(
			"• **%s -> %s:** `%s %s`",
			debt.OwedByUser.Name,
			debt.OwedToUser.Name,
			debt.Amount.StringFixed(2),
			strings.ToUpper(strings.TrimSpace(debt.Currency)),
		))
	}

	if len(debts) > maxRows {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("_Showing first %d of %d debts._", maxRows, len(debts)))
	}

	return strings.Join(lines, "\n")
}

func handleAddPaymentModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := extractModalValues(i.ModalSubmitData())
	currency := normalizePaymentCurrency(values["currency"])

	amount, err := strconv.ParseFloat(strings.TrimSpace(values["amount"]), 64)
	if err != nil {
		return respondWithMessage(s, i, fmt.Sprintf("invalid amount: %s", err.Error()))
	}

	debtors := splitCSV(values["debtors"])

	req := apimodels.Payment{
		Amount:           amount,
		Payer:            strings.TrimSpace(values["payer"]),
		Description:      strings.TrimSpace(values["description"]),
		Date:             time.Now().UTC(),
		FromExchangeRate: currency,
		ToExchangeRate:   "USD",
		Debtors:          debtors,
		DebtMode:         "equal",
	}

	if err := createPayment(&req); err != nil {
		return respondWithEmbed(s, i, paymentCreatedEmbed(req, err))
	}

	return respondWithEmbed(s, i, paymentCreatedEmbed(req, nil))
}

func handleGetPayments(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := respondDeferred(s, i); err != nil {
		return fmt.Errorf("defer getpayments response: %w", err)
	}

	payments, err := getPaymentResponses()
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("fetch payments failed: " + err.Error()),
		})
		if editErr != nil {
			return fmt.Errorf("fetch payments failed: %w; edit response failed: %w", err, editErr)
		}
		return nil
	}

	// Clear prev threads
	threads, err := s.ThreadsActive(i.ChannelID)
	if err != nil {
		return fmt.Errorf("obtain previous payment threads")
	}

	for _, thread := range threads.Threads {
		if thread.ParentID != i.ChannelID {
			continue
		}

		if thread.Name != "Payments" {
			continue
		}

		if _, err := s.ChannelDelete(thread.ID); err != nil {
			return fmt.Errorf("delete thread %s: %w", thread.ID, err)
		}
	}

	thread, err := createPaymentsThread(s, i.ChannelID, payments)
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("build payments thread failed: " + err.Error()),
		})
		if editErr != nil {
			return fmt.Errorf("build payments thread failed: %w; edit response failed: %w", err, editErr)
		}
		return nil
	}

	content := fmt.Sprintf("Posted %d payments in <#%s>.", len(payments), thread.ID)
	if len(payments) == 0 {
		content = fmt.Sprintf("Created <#%s>. No payments have been recorded yet.", thread.ID)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(content),
	})
	if err != nil {
		return fmt.Errorf("edit getpayments response: %w", err)
	}

	return nil
}

func handleGetDebts(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	debts, err := getDebtResponses()
	if err != nil {
		return respondWithMessage(s, i, "fetch debts failed: "+err.Error())
	}

	return respondWithMessage(s, i, debtListMessage(debts))
}

func paymentDeleteEmbed(paymentId uint, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{
			Title:       "Payment Deletion Failed",
			Description: fmt.Sprintf("An error occurred while recording the payment:\n%s", err.Error()),
			Color:       0xE74C3C,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "Payment Deleted",
		Description: "The payment was deleted successfully.",
		Color:       0x2ECC71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Id",
				Value:  "`" + strconv.FormatUint(uint64(paymentId), 10) + "`",
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func handleDeletePayment(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var paymentID int64
	for _, opt := range data.Options {
		if opt.Name == "id" {
			paymentID = opt.IntValue()
			break
		}
	}

	coercedPaymentId := uint(paymentID)

	// Do the deletion
	err := deletePayment(coercedPaymentId)
	embed := paymentDeleteEmbed(coercedPaymentId, err)

	return respondWithEmbed(s, i, embed)
}

func createUser(username string) error {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return fmt.Errorf("missing required env var: API_URL")
	}

	url := fmt.Sprintf("%s/api/v1/user", apiURL)

	apiPayload := apimodels.User{
		Name: username,
	}

	marshaledJson, err := json.Marshal(apiPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(marshaledJson))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func createUserEmbed(username string, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{
			Title:       "User Creation Failed",
			Description: fmt.Sprintf("An error occurred while creating the user:\n%s", err.Error()),
			Color:       0xE74C3C,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "User Created",
		Description: "The user was created successfully.",
		Color:       0x2ECC71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  "`" + username + "`",
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func createSettleUpEmbed(username string, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{
			Title:       "Settle Up Failed",
			Description: fmt.Sprintf("An error occurred while settling up for user:\n%s", err.Error()),
			Color:       0xE74C3C,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "Settled Up",
		Description: "You're all settled up!",
		Color:       0x2ECC71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  "`" + username + "`",
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func handleCreateUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var username string
	for _, opt := range data.Options {
		if opt.Name == "name" {
			username = opt.StringValue()
			break
		}
	}

	err := createUser(username)
	embed := createUserEmbed(username, err)

	return respondWithEmbed(s, i, embed)
}

func settleUp(username string) error {
	return nil
}

func handleSettle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var username string
	for _, opt := range data.Options {
		if opt.Name == "name" {
			username = opt.StringValue()
			break
		}
	}

	err := settleUp(username)
	embed := createSettleUpEmbed(username, err)

	if err := respondWithEmbed(s, i, embed); err != nil {
		return err
	}

	// So proud of you
	s.ChannelMessageSend(i.ChannelID, "So proud of you!")

	num := rand.IntN(6)

	switch num {
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

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		if data.Name == "payment" {
			if err := openAddPaymentModal(s, i); err != nil {
				log.Printf("open modal failed: %v", err)
			}
		}
		if data.Name == "getpayments" {
			if err := handleGetPayments(s, i); err != nil {
				log.Printf("getpayments failed: %v", err)
			}
		}
		if data.Name == "getdebts" {
			if err := handleGetDebts(s, i); err != nil {
				log.Printf("getdebts failed: %v", err)
			}
		}
		if data.Name == "delpayment" {
			if err := handleDeletePayment(s, i); err != nil {
				log.Printf("delpayment failed: %v", err)
			}
		}
		if data.Name == "adduser" {
			if err := handleCreateUser(s, i); err != nil {
				log.Printf("adduser failed: %v", err)
			}
		}
		if data.Name == "settle" {
			if err := handleSettle(s, i); err != nil {
				log.Printf("settle failed: %v", err)
			}
		}

	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		if data.CustomID == "addpayment_modal" {
			if err := handleAddPaymentModalSubmit(s, i); err != nil {
				log.Printf("modal submit failed: %v", err)
			}
		}
	}
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	parts := strings.Fields(m.Content[1:])
	if len(parts) == 0 {
		return
	}

	log.Printf("I see: %s", m.Content)

	command, _ := parts[0], parts[1:]

	switch command {
	case "fetch":
		body, err := getPayments()
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "fetch failed: "+err.Error())
			return
		}

		// Keep under Discord message size limits for quick testing.
		if len(body) > 1800 {
			body = body[:1800] + "\n... (truncated)"
		}

		_, _ = s.ChannelMessageSend(m.ChannelID, "```json\n"+body+"\n```")
	}
}
