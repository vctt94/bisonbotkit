package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

var (
	flagAppRoot = flag.String("approot", "~/.chatbot", "Path to application data directory")
)

func realMain() error {
	flag.Parse()

	// Load bot configuration
	cfg, err := config.LoadBotConfig(*flagAppRoot, "chatbot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create a bidirectional channel
	pmChan := make(chan *types.ReceivedPM)
	// Assign the send side to the config
	cfg.PMChan = pmChan
	// Note: PMLog will be created internally by NewBot

	// Create new bot instance
	bot, err := kit.NewBot(cfg)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}
	log := bot.LogBackend.Logger("ChatBot")
	cfg.PMLog = bot.LogBackend.Logger("PM")

	// Add a goroutine to handle PMs using our bidirectional channel
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg)
		}
	}()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add input handling goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			tokens := strings.SplitN(line, " ", 2)
			if len(tokens) != 2 {
				log.Warn("Invalid format. Use: <nick> <message>")
				continue
			}

			nick, msg := tokens[0], tokens[1]
			if err := bot.SendPM(ctx, nick, msg); err != nil {
				log.Warnf("Failed to send PM: %v", err)
				continue
			}
			log.Infof("-> %s: %s", nick, msg)
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("Error reading input: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received shutdown signal: %v", sig)
		bot.Close()
		cancel()
	}()

	// Run the bot with the cancellable context
	if err := bot.Run(ctx); err != nil {
		return fmt.Errorf("bot error: %v", err)
	}

	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
