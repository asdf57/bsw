package router

import (
	"log"
	"strings"

	"github.com/asdf57/bsw/cmd/bot/features/debt"
	"github.com/asdf57/bsw/cmd/bot/features/payment"
	"github.com/asdf57/bsw/cmd/bot/features/system"
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
			if err := payment.OpenPaymentCurrencyPicker(s, i); err != nil {
				log.Printf("open payment currency picker failed: %v", err)
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
		case "getsettlements":
			if err := debt.HandleSettlements(s, i); err != nil {
				log.Printf("getsettlements failed: %v", err)
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
		case "settleall":
			if err := debt.OpenSettleAllModal(s, i); err != nil {
				log.Printf("settleall failed: %v", err)
			}
		case "settlecustom":
			if err := debt.OpenSettleCustomStart(s, i); err != nil {
				log.Printf("settlecustom failed: %v", err)
			}
		case "reversesettlement":
			if err := debt.HandleReverseSettlement(s, i); err != nil {
				log.Printf("reversesettlement failed: %v", err)
			}
		case "exportdata":
			if err := system.HandleExportData(s, i); err != nil {
				log.Printf("exportdata failed: %v", err)
			}
		case "importdata":
			if err := system.HandleImportData(s, i); err != nil {
				log.Printf("importdata failed: %v", err)
			}
		}
	case discordgo.InteractionMessageComponent:
		data := i.MessageComponentData()
		switch {
		case data.CustomID == debt.SettleCustomOwedBySelectID:
			if err := debt.HandleSettleCustomOwedBySelected(s, i); err != nil {
				log.Printf("settle owed-by select failed: %v", err)
			}
		case strings.HasPrefix(data.CustomID, debt.SettleCustomOwedToPrefix):
			if err := debt.HandleSettleCustomOwedToSelected(s, i); err != nil {
				log.Printf("settle owed-to select failed: %v", err)
			}
		case data.CustomID == payment.CurrencySelectID:
			if err := payment.HandleCurrencySelected(s, i); err != nil {
				log.Printf("payment currency select failed: %v", err)
			}
		case strings.HasPrefix(data.CustomID, payment.PayerSelectID):
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
		case strings.HasPrefix(data.CustomID, debt.SettleAmountPrefix):
			if err := debt.HandleSettleAmountModalSubmit(s, i); err != nil {
				log.Printf("settle amount modal submit failed: %v", err)
			}
		case data.CustomID == debt.SettleAllModalID:
			if err := debt.HandleSettleAllModalSubmit(s, i); err != nil {
				log.Printf("settleall modal submit failed: %v", err)
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
