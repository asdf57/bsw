package tag

import (
	"fmt"
	"time"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/bwmarrin/discordgo"
)

func CreateEmbed(requestedName string, created *apimodels.TagResponse, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "Tag Creation Failed", Description: fmt.Sprintf("An error occurred while creating the tag:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	name := requestedName
	if created != nil && created.Name != "" {
		name = created.Name
	}
	return &discordgo.MessageEmbed{Title: "Tag Created", Description: "The tag is available for new payments.", Color: 0x2ECC71, Fields: []*discordgo.MessageEmbedField{{Name: "Tag", Value: "`" + name + "`", Inline: true}}, Timestamp: time.Now().Format(time.RFC3339)}
}
