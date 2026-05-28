package user

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func CreateEmbed(username string, discordHandle string, err error) *discordgo.MessageEmbed {
	if err != nil {
		return &discordgo.MessageEmbed{Title: "User Creation Failed", Description: fmt.Sprintf("An error occurred while creating the user:\n%s", err.Error()), Color: 0xE74C3C, Timestamp: time.Now().Format(time.RFC3339)}
	}
	fields := []*discordgo.MessageEmbedField{{Name: "User", Value: "`" + username + "`", Inline: true}}
	if discordHandle != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Discord", Value: "`" + discordHandle + "`", Inline: true})
	}
	return &discordgo.MessageEmbed{Title: "User Created", Description: "The user was created successfully.", Color: 0x2ECC71, Fields: fields, Timestamp: time.Now().Format(time.RFC3339)}
}
