package user

import (
	"strings"

	"github.com/asdf57/bsw/cmd/bot/shared"
	"github.com/bwmarrin/discordgo"
)

const (
	AddUserModalID       = "adduser_modal"
	AddUserNameInputID   = "adduser_name"
	AddUserHandleInputID = "adduser_discord_user"
)

func OpenCreateUserModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true
	optional := false

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: AddUserModalID,
			Title:    "Add User",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Name", Component: discordgo.TextInput{CustomID: AddUserNameInputID, Style: discordgo.TextInputShort, Placeholder: "Alice", Required: &required}},
				discordgo.Label{Label: "Discord User", Description: "Optional.", Component: discordgo.SelectMenu{MenuType: discordgo.UserSelectMenu, CustomID: AddUserHandleInputID, Placeholder: "Select Discord user", MaxValues: 1, Required: &optional}},
			},
		},
	})
}

func HandleCreateUserModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	username := strings.TrimSpace(values[AddUserNameInputID])
	discordHandle := discordMentionFromSelectedUser(values[AddUserHandleInputID])
	if username == "" {
		return shared.RespondWithMessage(s, i, "Name is required.")
	}

	err := shared.CreateUser(username, discordHandle)
	return shared.RespondWithEmbed(s, i, CreateEmbed(username, discordHandle, err))
}

func discordMentionFromSelectedUser(userID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	return "<@" + userID + ">"
}
