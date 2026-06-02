package payment

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

func DeleteEmbed(paymentID uint, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Payment Deletion Failed", Description: fmt.Sprintf("An error occurred while recording the payment:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	return &discordgo.MessageEmbed{Title: "Payment Deleted", Description: "The payment was deleted successfully.", Color: 0x2ECC71, Fields: []*discordgo.MessageEmbedField{{Name: "Id", Value: "`" + strconv.FormatUint(uint64(paymentID), 10) + "`", Inline: true}}, Timestamp: time.Now().Format(time.RFC3339)}
}

func CreatedEmbed(payment apimodels.Payment, paymentID uint, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Payment Creation Failed", Description: fmt.Sprintf("An error occurred while recording the payment:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	description := strings.TrimSpace(payment.Description)
	if description == "" {
		description = "_none_"
	}
	return &discordgo.MessageEmbed{Title: "Payment Created", Description: "The payment was recorded successfully.", Color: 0x2ECC71, Fields: []*discordgo.MessageEmbedField{{Name: "ID", Value: "`" + strconv.FormatUint(uint64(paymentID), 10) + "`", Inline: true}, {Name: "Amount", Value: "`" + shared.FormatAmount(payment) + "`", Inline: true}, {Name: "Date", Value: payment.Date.UTC().Format("2006-01-02 15:04 UTC"), Inline: true}, {Name: "Payer", Value: payment.Payer, Inline: true}, {Name: "Debtors", Value: formatDebtors(payment.Debtors), Inline: false}, {Name: "Tags", Value: formatTags(payment.Tags), Inline: false}, {Name: "Description", Value: description, Inline: false}}, Timestamp: time.Now().Format(time.RFC3339)}
}

func CreatedPaymentEmbed(payment apimodels.Payment, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Failed to fetch payment", Description: fmt.Sprintf("An error occurred while fetching the payment:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	description := strings.TrimSpace(payment.Description)
	if description == "" {
		description = "_none_"
	}
	return &discordgo.MessageEmbed{Title: "Payment", Description: "Payment stuff.", Color: 0x2ECC71, Fields: []*discordgo.MessageEmbedField{{Name: "Amount", Value: "`" + shared.FormatAmount(payment) + "`", Inline: true}, {Name: "Payer", Value: payment.Payer, Inline: true}, {Name: "Debtors", Value: formatDebtors(payment.Debtors), Inline: false}, {Name: "Description", Value: description, Inline: false}}, Timestamp: time.Now().Format(time.RFC3339)}
}

func threadEmbed(payment apimodels.PaymentResponse) *discordgo.MessageEmbed {
	description := strings.TrimSpace(payment.Description)
	if description == "" {
		description = "_none_"
	}
	fields := []*discordgo.MessageEmbedField{
		{Name: "Payer", Value: payment.Payer.Name, Inline: true},
		{Name: "Date", Value: payment.Date.Local().Format("2006-01-02 15:04"), Inline: true},
		{Name: "Debtors", Value: formatResponseDebtors(payment.Debtors), Inline: false},
		{Name: "Tags", Value: formatTags(payment.Tags), Inline: false},
		{Name: "Description", Value: description, Inline: false},
	}
	// if payment.Exchange.FromCurrency != "" && payment.Exchange.ToCurrency != "" && payment.Exchange.FromCurrency != payment.Exchange.ToCurrency {
	// 	fields = append(fields, &discordgo.MessageEmbedField{Name: "Exchange Rate", Value: fmt.Sprintf("1 %s = %.6f %s", payment.Exchange.FromCurrency, payment.Exchange.Rate, payment.Exchange.ToCurrency), Inline: false})
	// }
	return &discordgo.MessageEmbed{Title: fmt.Sprintf("#%d • %s", payment.ID, formatResponseAmount(payment)), Color: 0x3B82F6, Fields: fields, Timestamp: payment.Date.Format(time.RFC3339)}
}

// Build bulleted list of the debtors
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

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "general"
	}
	return strings.Join(tags, ", ")
}

func formatResponseAmount(payment apimodels.PaymentResponse) string {
	return shared.FormatDecimalAmount(payment.Amount, payment.Currency)
}

func formatResponseDebtors(debtors []apimodels.UserSummary) string {
	if len(debtors) == 0 {
		return ""
	}
	names := make([]string, 0, len(debtors))
	for _, debtor := range debtors {
		names = append(names, debtor.Name)
	}
	return strings.Join(names, ", ")
}

func ListTable(payments []apimodels.PaymentResponse) string {
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
		fmt.Sprintf("%s  %s  %s  %s  %s  %s", shared.PadCell("ID", idWidth), shared.PadCell("PAYER", payerWidth), shared.PadCell("DEBTORS", debtorsWidth), shared.PadCell("AMOUNT", amountWidth), shared.PadCell("DATE", dateWidth), shared.PadCell("DESCRIPTION", descriptionWidth)),
		fmt.Sprintf("%s  %s  %s  %s  %s  %s", strings.Repeat("-", idWidth), strings.Repeat("-", payerWidth), strings.Repeat("-", debtorsWidth), strings.Repeat("-", amountWidth), strings.Repeat("-", dateWidth), strings.Repeat("-", descriptionWidth)),
	}
	limit := shared.Min(len(payments), maxRows)
	for _, payment := range payments[:limit] {
		dateValue := payment.Date.Local().Format("2006-01-02 15:04")
		description := payment.Description
		if strings.TrimSpace(description) == "" {
			description = "-"
		}
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s  %s  %s", shared.PadCell(strconv.FormatUint(uint64(payment.ID), 10), idWidth), shared.PadCell(payment.Payer.Name, payerWidth), shared.PadCell(formatResponseDebtors(payment.Debtors), debtorsWidth), shared.PadCell(formatResponseAmount(payment), amountWidth), shared.PadCell(dateValue, dateWidth), shared.PadCell(description, descriptionWidth)))
	}
	if len(payments) > maxRows {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Showing first %d of %d payments.", maxRows, len(payments)))
	}
	return "```text\n" + strings.Join(lines, "\n") + "\n```"
}
