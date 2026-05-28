package payment

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/asdf57/bsw/cmd/bot/shared"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

const maxEmbedsPerMessage = 10

func currencyOptions(selected string) []discordgo.SelectMenuOption {
	selected = strings.ToUpper(strings.TrimSpace(selected))
	currencies := []string{"USD", "EUR", "GBP", "CAD", "JPY"}
	options := make([]discordgo.SelectMenuOption, 0, len(currencies))
	for _, currency := range currencies {
		options = append(options, discordgo.SelectMenuOption{Label: currency, Value: currency, Default: currency == selected})
	}
	if selected == "" {
		options[0].Default = true
	}
	return options
}

func OpenAddPaymentModal(s *discordgo.Session, i *discordgo.InteractionCreate, payer string) error {
	required := true
	minValues := 1
	optional := false

	users, err := shared.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	debtorOptions := userOptions(users, payer)
	tags, err := shared.GetTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}
	tagOpts := tagOptions(tags)
	maxTagValues := len(tagOpts)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: ModalCustomID(payer),
			Title:    "Add Payment",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Amount", Component: discordgo.TextInput{CustomID: "amount", Style: discordgo.TextInputShort, Placeholder: "10.00", Required: &required}},
				discordgo.Label{Label: "Description", Component: discordgo.TextInput{CustomID: "description", Style: discordgo.TextInputShort, Placeholder: "Dinner", Required: &required}},
				discordgo.Label{Label: "Tags", Description: "Select one or more. Leave empty for general.", Component: func() discordgo.SelectMenu {
					menu := discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "tags", Placeholder: "Select tags", Required: &optional, Options: tagOpts}
					if len(tagOpts) > 0 {
						menu.MaxValues = maxTagValues
					} else {
						menu.Disabled = true
						menu.Placeholder = "No tags available"
					}
					return menu
				}()},
				discordgo.Label{Label: "Currency", Description: "Choose the payment currency.", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "currency", Placeholder: "Select a currency", MaxValues: 1, Required: &required, Options: currencyOptions("USD")}},
				discordgo.Label{Label: "Debtors", Component: func() discordgo.SelectMenu {
					menu := discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "debtors", Placeholder: "Select debtors", Required: &optional, Options: debtorOptions}
					if len(debtorOptions) > 0 {
						menu.MinValues = &minValues
						menu.MaxValues = len(debtorOptions)
					} else {
						menu.Disabled = true
						menu.Placeholder = "No eligible debtors for selected payer"
					}
					return menu
				}()},
			},
		},
	})
}

func OpenPaymentPayerPicker(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	users, err := shared.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	userOpts := userOptions(users, "")
	if len(userOpts) == 0 {
		return shared.RespondWithMessage(s, i, "No users available. Add users first with `/adduser`.")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select a payer to continue.",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: PayerSelectID, Placeholder: "Select payer", MaxValues: 1, Required: &required, Options: userOpts}}},
			},
		},
	})
}

func OpenEditPaymentModal(s *discordgo.Session, i *discordgo.InteractionCreate, paymentID uint) error {
	required := true
	optional := false

	payment, err := shared.GetPaymentResponse(paymentID)
	if err != nil {
		return shared.RespondWithMessage(s, i, "fetch payment failed: "+err.Error())
	}
	users, err := shared.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	tags, err := shared.GetTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	debtorOpts := selectedUserOptions(users, payment.Payer.Name, payment.Debtors)
	tagOpts := selectedTagOptions(tags, payment.Tags)
	maxDebtors := len(debtorOpts)
	maxTags := len(tagOpts)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: EditModalCustomID(paymentID),
			Title:    "Edit Payment",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Amount", Component: discordgo.TextInput{CustomID: "amount", Style: discordgo.TextInputShort, Value: payment.Amount.StringFixed(2), Required: &required}},
				discordgo.Label{Label: "Description", Component: discordgo.TextInput{CustomID: "description", Style: discordgo.TextInputShort, Value: payment.Description, Required: &required}},
				discordgo.Label{Label: "Tags", Description: "Select one or more. Leave empty for general.", Component: func() discordgo.SelectMenu {
					menu := discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "tags", Placeholder: "Select tags", Required: &optional, Options: tagOpts}
					if len(tagOpts) > 0 {
						menu.MaxValues = maxTags
					} else {
						menu.Disabled = true
						menu.Placeholder = "No tags available"
					}
					return menu
				}()},
				discordgo.Label{Label: "Currency", Description: "Choose the payment currency.", Component: discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "currency", Placeholder: "Select a currency", MaxValues: 1, Required: &required, Options: currencyOptions(payment.Currency)}},
				discordgo.Label{Label: "Debtors", Component: func() discordgo.SelectMenu {
					menu := discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: "debtors", Placeholder: "Select debtors", Required: &optional, Options: debtorOpts}
					if len(debtorOpts) > 0 {
						menu.MaxValues = maxDebtors
					} else {
						menu.Disabled = true
						menu.Placeholder = "No eligible debtors for selected payer"
					}
					return menu
				}()},
			},
		},
	})
}

func OpenPaymentRangePicker(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	optional := false
	tags, err := shared.GetTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}
	tagOpts := tagOptions(tags)
	maxTagValues := len(tagOpts)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: PaymentRangeModalID,
			Title:    "Get Payments",
			Components: []discordgo.MessageComponent{
				discordgo.Label{
					Label: "Payment Range",
					Component: discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						CustomID:    PaymentRangeSelectID,
						Placeholder: "Select range",
						MaxValues:   1,
						Required:    &required,
						Options: []discordgo.SelectMenuOption{
							{Label: "All", Value: PaymentRangeAll},
							{Label: "Today", Value: PaymentRangeToday},
							{Label: "Yesterday", Value: PaymentRangeYesterday},
							{Label: "Last 7 days", Value: PaymentRangeLast7Days},
							{Label: "Custom", Value: PaymentRangeCustom},
						},
					},
				},
				discordgo.Label{
					Label:       "Tag Operation",
					Description: "Used when tags are selected.",
					Component: discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						CustomID:    PaymentTagOpSelectID,
						Placeholder: "Select tag operation",
						MaxValues:   1,
						Required:    &optional,
						Options: []discordgo.SelectMenuOption{
							{Label: "All selected tags (AND)", Value: PaymentTagOpAnd, Default: true},
							{Label: "Any selected tag (OR)", Value: PaymentTagOpOr},
						},
					},
				},
				discordgo.Label{
					Label:       "Tags",
					Description: "Select one or more tags to filter payments.",
					Component: func() discordgo.SelectMenu {
						menu := discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: PaymentTagsSelectID, Placeholder: "Select tags", Required: &optional, Options: tagOpts}
						if len(tagOpts) > 0 {
							menu.MaxValues = maxTagValues
						} else {
							menu.Disabled = true
							menu.Placeholder = "No tags available"
						}
						return menu
					}(),
				},
				discordgo.Label{Label: "Custom Since", Description: "Only used when Custom is selected. Use YYYY-MM-DD, MM/DD/YYYY, or 10d.", Component: discordgo.TextInput{CustomID: CustomRangeInputID, Style: discordgo.TextInputShort, Placeholder: "2026-05-01", Required: &optional}},
			},
		},
	})
}

func CreatePaymentsThread(s *discordgo.Session, channelID string, payments []apimodels.PaymentResponse) (*discordgo.Channel, error) {
	totalStart := time.Now()
	log.Printf("getpayments: CreatePaymentsThread start channel=%s payments=%d", channelID, len(payments))

	threadStart := time.Now()
	thread, err := s.ThreadStart(channelID, "Payments", discordgo.ChannelTypeGuildPublicThread, 1440)
	if err != nil {
		return nil, fmt.Errorf("start payments thread: %w", err)
	}
	log.Printf("getpayments: started thread id=%s in %s", thread.ID, time.Since(threadStart))

	messagesStart := time.Now()
	messages, err := s.ChannelMessages(channelID, 10, "", "", "")
	if err != nil {
		return nil, err
	}
	log.Printf("getpayments: fetched %d recent channel messages in %s", len(messages), time.Since(messagesStart))

	deletedMessages := 0
	for _, msg := range messages {
		if msg.Type == discordgo.MessageTypeThreadCreated {
			deleteStart := time.Now()
			if err := s.ChannelMessageDelete(channelID, msg.ID); err != nil {
				return nil, err
			}
			deletedMessages++
			log.Printf("getpayments: deleted thread-created message id=%s in %s", msg.ID, time.Since(deleteStart))
			continue
		}
		if msg.Author != nil && msg.Author.ID == s.State.User.ID && strings.HasPrefix(msg.Content, "Posted ") && strings.Contains(msg.Content, " payments in ") {
			deleteStart := time.Now()
			if err := s.ChannelMessageDelete(channelID, msg.ID); err != nil {
				return nil, err
			}
			deletedMessages++
			log.Printf("getpayments: deleted previous summary message id=%s in %s", msg.ID, time.Since(deleteStart))
		}
	}
	log.Printf("getpayments: deleted %d cleanup messages", deletedMessages)

	if len(payments) == 0 {
		sendStart := time.Now()
		if _, err := s.ChannelMessageSend(thread.ID, "No payments found."); err != nil {
			return nil, fmt.Errorf("post empty payments message: %w", err)
		}
		log.Printf("getpayments: posted empty message in %s", time.Since(sendStart))
		log.Printf("getpayments: CreatePaymentsThread done in %s", time.Since(totalStart))
		return thread, nil
	}

	batches := paymentEmbedBatches(payments, maxEmbedsPerMessage)
	postStart := time.Now()
	for idx, batch := range batches {
		batchStart := time.Now()
		if _, err := s.ChannelMessageSendEmbeds(thread.ID, batch); err != nil {
			return nil, fmt.Errorf("post payment embed batch %d of %d to thread: %w", idx+1, len(batches), err)
		}
		log.Printf("getpayments: posted payment embed batch %d/%d embeds=%d in %s", idx+1, len(batches), len(batch), time.Since(batchStart))
	}
	log.Printf("getpayments: posted %d payment embeds in %d batches in %s", len(payments), len(batches), time.Since(postStart))
	log.Printf("getpayments: CreatePaymentsThread done in %s", time.Since(totalStart))

	return thread, nil
}

func paymentEmbedBatches(payments []apimodels.PaymentResponse, maxEmbeds int) [][]*discordgo.MessageEmbed {
	if len(payments) == 0 {
		return nil
	}
	if maxEmbeds <= 0 {
		maxEmbeds = maxEmbedsPerMessage
	}

	batches := make([][]*discordgo.MessageEmbed, int(math.Ceil(float64(len(payments))/float64(maxEmbeds))))
	start := 0
	for batchIdx := range batches {
		end := start + maxEmbeds
		if end > len(payments) {
			end = len(payments)
		}

		for i := start; i < end; i++ {
			batches[batchIdx] = append(batches[batchIdx], threadEmbed(payments[i]))
		}
		start = end
	}
	return batches
}
