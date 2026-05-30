package router

import "github.com/bwmarrin/discordgo"

func RegisterCommands(s *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{Name: "payment", Description: "Create a new payment"},
		{Name: "addtag", Description: "Create a payment tag"},
		{Name: "getpayments", Description: "Show all recorded payments"},
		{Name: "getdebts", Description: "Show all current debts", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "currency", Description: "Currency to generate debts in", Required: false},
		}},
		{Name: "delpayment", Description: "Delete a payment", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "id", Description: "Payment ID", Required: true},
		}},
		{Name: "editpayment", Description: "Edit a payment", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "id", Description: "Payment ID", Required: true},
		}},
		{Name: "adduser", Description: "Create a user"},
		{Name: "settle", Description: "Settle up"},
		{Name: "stats", Description: "Show user spending stats"},
		{Name: "settlements", Description: "Show settlement history"},
		{Name: "reversesettlement", Description: "Reverse a settlement", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "id", Description: "Settlement ID", Required: true},
		}},
		{Name: "exportdata", Description: "Export a system checkpoint JSON file"},
		{Name: "importdata", Description: "Import a system checkpoint JSON file", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionAttachment, Name: "file", Description: "Checkpoint JSON file", Required: true},
		}},
	}
	for _, cmd := range commands {
		if _, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd); err != nil {
			return err
		}
	}
	return nil
}
