package payment

import (
	"net/url"
	"strconv"
	"strings"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

const (
	PayerSelectID         = "payment_select_payer"
	PaymentRangeSelectID  = "payment_select_range"
	PaymentTagsSelectID   = "payment_select_tags"
	PaymentTagOpSelectID  = "payment_select_tag_op"
	PaymentRangeAll       = "all"
	PaymentRangeToday     = "today"
	PaymentRangeYesterday = "yesterday"
	PaymentRangeLast7Days = "last_7_days"
	PaymentRangeCustom    = "custom"
	PaymentTagOpAnd       = "and"
	PaymentTagOpOr        = "or"
	AddModalPrefix        = "addpayment_modal:payer="
	EditModalPrefix       = "editpayment_modal:id="
	PaymentRangeModalID   = "getpayments_range_modal"
	CustomRangeInputID    = "payment_range_since"
)

func ModalCustomID(payer string) string {
	return AddModalPrefix + url.QueryEscape(strings.TrimSpace(payer))
}

func PayerFromModalCustomID(customID string) (string, bool) {
	if !strings.HasPrefix(customID, AddModalPrefix) {
		return "", false
	}
	raw := strings.TrimPrefix(customID, AddModalPrefix)
	payer, err := url.QueryUnescape(raw)
	if err != nil {
		return "", false
	}
	payer = strings.TrimSpace(payer)
	if payer == "" {
		return "", false
	}
	return payer, true
}

func EditModalCustomID(paymentID uint) string {
	return EditModalPrefix + url.QueryEscape(strconv.FormatUint(uint64(paymentID), 10))
}

func PaymentIDFromEditModalCustomID(customID string) (uint, bool) {
	if !strings.HasPrefix(customID, EditModalPrefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(customID, EditModalPrefix)
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return 0, false
	}
	id, err := strconv.ParseUint(strings.TrimSpace(decoded), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func userOptions(users []apimodels.UserSummary, exclude string) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, len(users))
	for _, u := range users {
		name := strings.TrimSpace(u.Name)
		if name == "" || strings.EqualFold(name, exclude) {
			continue
		}
		options = append(options, discordgo.SelectMenuOption{Label: name, Value: name})
	}
	return options
}

func tagOptions(tags []apimodels.TagResponse) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, len(tags))
	for _, tag := range tags {
		name := strings.ToLower(strings.TrimSpace(tag.Name))
		if name == "" {
			continue
		}
		options = append(options, discordgo.SelectMenuOption{Label: name, Value: name})
		if len(options) == 25 {
			break
		}
	}
	return options
}

func selectedTagOptions(tags []apimodels.TagResponse, selected []string) []discordgo.SelectMenuOption {
	selectedSet := make(map[string]struct{}, len(selected))
	for _, tag := range selected {
		selectedSet[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}

	options := tagOptions(tags)
	for idx := range options {
		if _, ok := selectedSet[strings.ToLower(options[idx].Value)]; ok {
			options[idx].Default = true
		}
	}
	return options
}

func selectedUserOptions(users []apimodels.UserSummary, exclude string, selected []apimodels.UserSummary) []discordgo.SelectMenuOption {
	selectedSet := make(map[string]struct{}, len(selected))
	for _, user := range selected {
		selectedSet[strings.TrimSpace(user.Name)] = struct{}{}
	}

	options := userOptions(users, exclude)
	for idx := range options {
		if _, ok := selectedSet[options[idx].Value]; ok {
			options[idx].Default = true
		}
	}
	return options
}
