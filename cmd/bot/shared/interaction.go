package shared

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ExtractModalValues(data discordgo.ModalSubmitInteractionData) map[string]string {
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

func SplitCSV(s string) []string {
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

func StringPtr(value string) *string { return &value }

func RespondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
}

func RespondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
}

func RespondDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
}

func PadCell(value string, width int) string {
	return fmt.Sprintf("%-*s", width, truncateCell(value, width))
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
