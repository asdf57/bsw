package payment

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

var relativeDaysPattern = regexp.MustCompile(`^(\d+)d$`)

func HandleGetPayments(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return OpenPaymentRangePicker(s, i)
}

func HandlePaymentRangeModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	rangeValue := strings.TrimSpace(values[PaymentRangeSelectID])
	if rangeValue == "" {
		return shared.RespondWithMessage(s, i, "Select a payment range.")
	}

	filter, err := paymentRangeFilterFromValues(rangeValue, strings.TrimSpace(values[CustomRangeInputID]), time.Now())
	if err != nil {
		return shared.RespondWithMessage(s, i, err.Error())
	}
	filter.tags = normalizeTags(shared.SplitCSV(values[PaymentTagsSelectID]))
	filter.tagOp = strings.ToLower(strings.TrimSpace(values[PaymentTagOpSelectID]))
	if filter.tagOp == "" {
		filter.tagOp = PaymentTagOpAnd
	}
	if filter.tagOp != PaymentTagOpAnd && filter.tagOp != PaymentTagOpOr {
		return shared.RespondWithMessage(s, i, "Select AND or OR for the tag operation.")
	}
	filter.label = filterLabel(filter)
	return respondWithFilteredPayments(s, i, filter)
}

func respondWithFilteredPayments(s *discordgo.Session, i *discordgo.InteractionCreate, filter paymentFilter) error {
	totalStart := time.Now()
	log.Printf("getpayments: start range=%q channel=%s", filter.label, i.ChannelID)

	deferStart := time.Now()
	if err := shared.RespondDeferred(s, i); err != nil {
		return fmt.Errorf("defer getpayments response: %w", err)
	}
	log.Printf("getpayments: deferred interaction in %s", time.Since(deferStart))

	fetchStart := time.Now()
	var payments []apimodels.PaymentResponse
	var err error
	if len(filter.tags) > 0 {
		payments, err = shared.GetPaymentResponsesByTags(filter.tags, filter.tagOp)
	} else {
		payments, err = shared.GetPaymentResponses()
	}
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: shared.StringPtr("fetch payments failed: " + err.Error())})
		if editErr != nil {
			return fmt.Errorf("fetch payments failed: %w; edit response failed: %w", err, editErr)
		}
		return nil
	}
	log.Printf("getpayments: fetched %d payments in %s", len(payments), time.Since(fetchStart))

	threadsStart := time.Now()
	threads, err := s.ThreadsActive(i.ChannelID)
	if err != nil {
		return fmt.Errorf("obtain previous payment threads")
	}
	log.Printf("getpayments: fetched %d active threads in %s", len(threads.Threads), time.Since(threadsStart))

	deletedThreads := 0
	for _, thread := range threads.Threads {
		if thread.ParentID != i.ChannelID || thread.Name != "Payments" {
			continue
		}
		deleteStart := time.Now()
		if _, err := s.ChannelDelete(thread.ID); err != nil {
			return fmt.Errorf("delete thread %s: %w", thread.ID, err)
		}
		deletedThreads++
		log.Printf("getpayments: deleted previous payments thread id=%s in %s", thread.ID, time.Since(deleteStart))
	}
	log.Printf("getpayments: deleted %d previous payments threads", deletedThreads)

	filterStart := time.Now()
	filteredPayments := filterPayments(payments, filter)
	log.Printf("getpayments: filtered %d payments down to %d in %s", len(payments), len(filteredPayments), time.Since(filterStart))

	var content string

	if len(filteredPayments) > 0 {
		threadStart := time.Now()
		thread, err := CreatePaymentsThread(s, i.ChannelID, filteredPayments)
		if err != nil {
			_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: shared.StringPtr("build payments thread failed: " + err.Error())})
			if editErr != nil {
				return fmt.Errorf("build payments thread failed: %w; edit response failed: %w", err, editErr)
			}
			return nil
		}
		log.Printf("getpayments: created/populated payments thread id=%s with %d payments in %s", thread.ID, len(filteredPayments), time.Since(threadStart))

		content = fmt.Sprintf("Posted %d payments in <#%s> for %s.", len(filteredPayments), thread.ID, filter.label)
	} else {
		content = fmt.Sprintf("No payments found for %s.", filter.label)
	}

	editStart := time.Now()
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: shared.StringPtr(content)})
	if err != nil {
		return fmt.Errorf("edit getpayments response: %w", err)
	}
	log.Printf("getpayments: edited interaction response in %s", time.Since(editStart))
	log.Printf("getpayments: done in %s", time.Since(totalStart))
	return nil
}

func paymentRangeFilterFromValues(rangeValue string, customInput string, now time.Time) (paymentFilter, error) {
	if rangeValue == PaymentRangeCustom {
		return customPaymentRangeFilter(customInput, now)
	}
	return paymentRangeFilter(rangeValue, now)
}

type paymentFilter struct {
	label string
	start *time.Time
	end   *time.Time
	tags  []string
	tagOp string
}

func paymentRangeFilter(value string, now time.Time) (paymentFilter, error) {
	today := startOfDay(now)
	switch value {
	case PaymentRangeAll:
		return paymentFilter{label: "all", start: nil}, nil
	case PaymentRangeToday:
		return paymentFilter{label: "today", start: &today}, nil
	case PaymentRangeYesterday:
		end := today
		start := today.AddDate(0, 0, -1)
		return paymentFilter{label: "yesterday", start: &start, end: &end}, nil
	case PaymentRangeLast7Days:
		start := today.AddDate(0, 0, -6)
		return paymentFilter{label: "the last 7 days", start: &start}, nil
	default:
		return paymentFilter{}, fmt.Errorf("unknown payment range: %s", value)
	}
}

func customPaymentRangeFilter(input string, now time.Time) (paymentFilter, error) {
	if input == "" {
		return paymentFilter{}, fmt.Errorf("enter a date as YYYY-MM-DD, MM/DD/YYYY, or a relative value like 10d")
	}

	cleaned := strings.ToLower(strings.TrimSpace(input))
	if matches := relativeDaysPattern.FindStringSubmatch(cleaned); len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err != nil || days <= 0 {
			return paymentFilter{}, fmt.Errorf("relative dates must use a positive day count, like 10d")
		}
		startDate := startOfDay(now).AddDate(0, 0, -(days - 1))
		return paymentFilter{label: fmt.Sprintf("the last %d days", days), start: &startDate}, nil
	}

	for _, layout := range []string{"2006-01-02", "01/02/2006"} {
		parsed, err := time.ParseInLocation(layout, cleaned, time.Local)
		if err == nil {
			start := startOfDay(parsed)
			return paymentFilter{label: "since " + start.Format("2006-01-02"), start: &start}, nil
		}
	}

	return paymentFilter{}, fmt.Errorf("invalid date range %q; use YYYY-MM-DD, MM/DD/YYYY, or a relative value like 10d", input)
}

func filterPayments(payments []apimodels.PaymentResponse, filter paymentFilter) []apimodels.PaymentResponse {
	if filter.start == nil && filter.end == nil {
		return payments
	}

	filtered := make([]apimodels.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentDate := payment.Date.In(time.Local)
		if filter.start != nil && paymentDate.Before(*filter.start) {
			continue
		}
		if filter.end != nil && !paymentDate.Before(*filter.end) {
			continue
		}
		filtered = append(filtered, payment)
	}
	return filtered
}

func filterLabel(filter paymentFilter) string {
	label := filter.label
	if label == "" {
		label = "all"
	}
	if len(filter.tags) == 0 {
		return label
	}
	op := strings.ToUpper(filter.tagOp)
	if op == "" {
		op = "AND"
	}
	return fmt.Sprintf("%s tagged %s", label, strings.Join(filter.tags, " "+op+" "))
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}

func startOfDay(t time.Time) time.Time {
	local := t.In(time.Local)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local)
}

func HandleAddPaymentModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	payer, ok := PayerFromModalCustomID(i.ModalSubmitData().CustomID)
	if !ok {
		return shared.RespondWithMessage(s, i, "invalid payment modal metadata; please run `/payment` again")
	}

	values := shared.ExtractModalValues(i.ModalSubmitData())
	currency := shared.NormalizePaymentCurrency(values["currency"])

	amount, err := strconv.ParseFloat(strings.TrimSpace(values["amount"]), 64)
	if err != nil {
		return shared.RespondWithMessage(s, i, fmt.Sprintf("invalid amount: %s", err.Error()))
	}
	if amount < 0 {
		return shared.RespondWithMessage(s, i, fmt.Sprintf("amount cannot be negative: %f", amount))
	}

	tags := normalizeTags(shared.SplitCSV(values["tags"]))
	debtors := shared.SplitCSV(values["debtors"])
	filtered := make([]string, 0, len(debtors))
	for _, debtor := range debtors {
		if strings.EqualFold(strings.TrimSpace(debtor), payer) {
			continue
		}
		filtered = append(filtered, debtor)
	}

	req := apimodels.Payment{Amount: amount, Payer: payer, Description: strings.TrimSpace(values["description"]), Date: time.Now().UTC(), Currency: currency, Debtors: filtered, Tags: tags, DebtMode: "equal"}
	if err := shared.CreatePayment(&req); err != nil {
		return shared.RespondWithEmbed(s, i, CreatedEmbed(req, err))
	}
	return shared.RespondWithEmbed(s, i, CreatedEmbed(req, nil))
}

func HandleEditPaymentCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var paymentID int64
	for _, opt := range data.Options {
		if opt.Name == "id" {
			paymentID = opt.IntValue()
			break
		}
	}
	if paymentID <= 0 {
		return shared.RespondWithMessage(s, i, "Provide a valid payment ID.")
	}
	return OpenEditPaymentModal(s, i, uint(paymentID))
}

func HandleEditPaymentModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	paymentID, ok := PaymentIDFromEditModalCustomID(i.ModalSubmitData().CustomID)
	if !ok {
		return shared.RespondWithMessage(s, i, "invalid payment edit metadata; please run `/editpayment` again")
	}

	existing, err := shared.GetPaymentResponse(paymentID)
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch payment failed: "+err.Error())
	}

	values := shared.ExtractModalValues(i.ModalSubmitData())
	amount, err := strconv.ParseFloat(strings.TrimSpace(values["amount"]), 64)
	if err != nil {
		return shared.RespondWithMessage(s, i, fmt.Sprintf("invalid amount: %s", err.Error()))
	}
	if amount < 0 {
		return shared.RespondWithMessage(s, i, fmt.Sprintf("amount cannot be negative: %f", amount))
	}

	tags := normalizeTags(shared.SplitCSV(values["tags"]))
	debtors := shared.SplitCSV(values["debtors"])
	filtered := make([]string, 0, len(debtors))
	for _, debtor := range debtors {
		if strings.EqualFold(strings.TrimSpace(debtor), existing.Payer.Name) {
			continue
		}
		filtered = append(filtered, debtor)
	}

	req := apimodels.Payment{
		Amount:      amount,
		Payer:       existing.Payer.Name,
		Description: strings.TrimSpace(values["description"]),
		Date:        existing.Date,
		Currency:    shared.NormalizePaymentCurrency(values["currency"]),
		Debtors:     filtered,
		Tags:        tags,
		DebtMode:    "equal",
	}
	updated, err := shared.UpdatePayment(paymentID, &req)
	if err != nil {
		return shared.RespondWithMessage(s, i, "update payment failed: "+err.Error())
	}
	return shared.RespondWithMessage(s, i, fmt.Sprintf("Updated payment #%d (%s).", updated.ID, formatResponseAmount(*updated)))
}

func HandleDeletePayment(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	var paymentID int64
	for _, opt := range data.Options {
		if opt.Name == "id" {
			paymentID = opt.IntValue()
			break
		}
	}
	coerced := uint(paymentID)
	err := shared.DeletePayment(coerced)
	return shared.RespondWithEmbed(s, i, DeleteEmbed(coerced, err))
}

func HandlePayerSelected(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	if data.CustomID != PayerSelectID {
		return nil
	}
	if len(data.Values) == 0 || strings.TrimSpace(data.Values[0]) == "" {
		return shared.RespondWithMessage(s, i, "Select a payer to continue.")
	}
	payer := strings.TrimSpace(data.Values[0])
	if err := OpenAddPaymentModal(s, i, payer); err != nil {
		return err
	}
	if i.Message != nil && i.ChannelID != "" && i.Message.ID != "" {
		if err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID); err != nil {
			log.Printf("delete payer prompt failed: %v", err)
		}
	}
	return nil
}
