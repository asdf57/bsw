package tag

import (
	"strings"

	"github.com/asdf57/bsw/cmd/bot/shared"
	"github.com/bwmarrin/discordgo"
)

const (
	AddTagModalID     = "addtag_modal"
	AddTagNameInputID = "addtag_name"
)

func OpenCreateTagModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	required := true

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: AddTagModalID,
			Title:    "Add Tag",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Label: "Tag", Component: discordgo.TextInput{CustomID: AddTagNameInputID, Style: discordgo.TextInputShort, Placeholder: "trip", Required: &required}},
			},
		},
	})
}

func HandleCreateTagModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	values := shared.ExtractModalValues(i.ModalSubmitData())
	name := strings.TrimSpace(values[AddTagNameInputID])
	if name == "" {
		return shared.RespondWithMessage(s, i, "Tag is required.")
	}

	created, err := shared.CreateTag(name)
	return shared.RespondWithEmbed(s, i, CreateEmbed(name, created, err))
}
