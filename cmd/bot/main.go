package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/asdf57/bsw/cmd/bot/router"
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

	dg.AddHandler(router.OnMessage)
	dg.AddHandler(router.OnInteractionCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	if err := dg.Open(); err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}

	if err := router.RegisterCommands(dg); err != nil {
		log.Fatalf("failed to register commands: %v", err)
	}

	log.Println("discord bot is connected; waiting for shutdown signal")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received; closing discord session")
}
