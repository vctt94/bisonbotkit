package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/utils"
)

func main() {
	// Get application data directory
	appdata := utils.AppDataDir("simplebot", false)

	// Load bot configuration
	cfg, err := config.LoadBotConfig(appdata, "simplebot.conf")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("Received shutdown signal\n")
		cancel()
	}()

	// Create and run the bot
	bot, err := bisonbotkit.NewBot(cfg)
	if err != nil {
		fmt.Printf("Failed to create bot: %v\n", err)
		os.Exit(1)
	}
	log := bot.LogBackend.Logger("SimpleBot")

	// Run the bot
	if err := bot.Run(ctx); err != nil {
		log.Errorf("Bot error: %v", err)
		os.Exit(1)
	}
}
