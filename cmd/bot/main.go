package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		log.Fatal("missing required env var: DISCORD_BOT_TOKEN")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("failed to create discord session: %v", err)
	}
	defer dg.Close()

	dg.AddHandler(onMessage)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	if err := dg.Open(); err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}

	log.Println("discord bot is connected; waiting for shutdown signal")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received; closing discord session")
}

func getPayments() (string, error) {
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	if apiURL == "" {
		return "", fmt.Errorf("missing required env var: API_URL")
	}

	resp, err := http.Get(apiURL + "/api/v1/payment/all")
	if err != nil {
		return "", fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body response: %w", err)
	}

	return string(body), nil
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	command, _ := parts[0], parts[1:]

	switch command {
	case "fetch":
		body, err := getPayments()
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "fetch failed: "+err.Error())
			return
		}

		// Keep under Discord message size limits for quick testing.
		if len(body) > 1800 {
			body = body[:1800] + "\n... (truncated)"
		}

		_, _ = s.ChannelMessageSend(m.ChannelID, "```json\n"+body+"\n```")
	}
}
