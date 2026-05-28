package api

type User struct {
	Name          string `json:"name"`
	DiscordHandle string `json:"discordHandle"`
}

type UserSummary struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	DiscordHandle string `json:"discordHandle"`
}
