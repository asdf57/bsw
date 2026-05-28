package router

import (
	"log"
	"strings"

	"github.com/asdf57/bsw/cmd/bot/features/debt"
	"github.com/asdf57/bsw/cmd/bot/features/payment"
	"github.com/asdf57/bsw/cmd/bot/features/tag"
	"github.com/asdf57/bsw/cmd/bot/features/user"
	"github.com/bwmarrin/discordgo"
)

func OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		switch data.Name {
		case "payment":
			if err := payment.OpenPaymentPayerPicker(s, i); err != nil {
				log.Printf("open modal failed: %v", err)
			}
		case "addtag":
			if err := tag.OpenCreateTagModal(s, i); err != nil {
				log.Printf("addtag failed: %v", err)
			}
		case "getpayments":
			if err := payment.HandleGetPayments(s, i); err != nil {
				log.Printf("getpayments failed: %v", err)
			}
		case "getdebts":
			if err := debt.HandleGetDebts(s, i); err != nil {
				log.Printf("getdebts failed: %v", err)
			}
		case "stats":
			if err := debt.OpenStatsModal(s, i); err != nil {
				log.Printf("stats failed: %v", err)
			}
		case "settlements":
			if err := debt.HandleSettlements(s, i); err != nil {
				log.Printf("settlements failed: %v", err)
			}
		case "delpayment":
			if err := payment.HandleDeletePayment(s, i); err != nil {
				log.Printf("delpayment failed: %v", err)
			}
		case "editpayment":
			if err := payment.HandleEditPaymentCommand(s, i); err != nil {
				log.Printf("editpayment failed: %v", err)
			}
		case "adduser":
			if err := user.OpenCreateUserModal(s, i); err != nil {
				log.Printf("adduser failed: %v", err)
			}
		case "settle":
			if err := debt.OpenSettleModal(s, i); err != nil {
				log.Printf("settle failed: %v", err)
			}
		case "reversesettlement":
			if err := debt.HandleReverseSettlement(s, i); err != nil {
				log.Printf("reversesettlement failed: %v", err)
			}
		}
	case discordgo.InteractionMessageComponent:
		data := i.MessageComponentData()
		switch data.CustomID {
		case payment.PayerSelectID:
			if err := payment.HandlePayerSelected(s, i); err != nil {
				log.Printf("payment payer select failed: %v", err)
			}
		}
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		switch {
		case strings.HasPrefix(data.CustomID, payment.AddModalPrefix):
			if err := payment.HandleAddPaymentModalSubmit(s, i); err != nil {
				log.Printf("modal submit failed: %v", err)
			}
		case strings.HasPrefix(data.CustomID, payment.EditModalPrefix):
			if err := payment.HandleEditPaymentModalSubmit(s, i); err != nil {
				log.Printf("edit payment modal submit failed: %v", err)
			}
		case data.CustomID == payment.PaymentRangeModalID:
			if err := payment.HandlePaymentRangeModalSubmit(s, i); err != nil {
				log.Printf("payment range modal submit failed: %v", err)
			}
		case data.CustomID == user.AddUserModalID:
			if err := user.HandleCreateUserModalSubmit(s, i); err != nil {
				log.Printf("adduser modal submit failed: %v", err)
			}
		case data.CustomID == tag.AddTagModalID:
			if err := tag.HandleCreateTagModalSubmit(s, i); err != nil {
				log.Printf("addtag modal submit failed: %v", err)
			}
		case data.CustomID == debt.SettleModalID:
			if err := debt.HandleSettleModalSubmit(s, i); err != nil {
				log.Printf("settle modal submit failed: %v", err)
			}
		case data.CustomID == debt.StatsModalID:
			if err := debt.HandleStatsModalSubmit(s, i); err != nil {
				log.Printf("stats modal submit failed: %v", err)
			}
		}
	}
}

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
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
}
