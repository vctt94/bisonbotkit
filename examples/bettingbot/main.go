package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/utils"
)

var (
	flagAppRoot = flag.String("approot", "~/.bettingbot", "Path to application data directory")
)

// handlePM handles incoming PM commands.
func handlePM(ctx context.Context, bot *kit.Bot, pm *types.ReceivedPM) {
	tokens := strings.Fields(pm.Msg.Message)
	if len(tokens) == 0 {
		return
	}

	cmd := strings.ToLower(tokens[0])

	// Expected usage: "bet <amount in DCR> <odd|even>"
	if cmd == "bet" && len(tokens) == 3 {
		// 1) Parse the bet amount
		betFloat, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			bot.SendPM(ctx, pm.Nick, "Invalid bet amount. Please enter a valid number.")
			return
		}

		// Convert float to dcrutil.Amount
		betAmount, err := dcrutil.NewAmount(betFloat)
		if err != nil {
			bot.SendPM(ctx, pm.Nick, "Invalid DCR amount. Please enter a valid number.")
			return
		}
		if betAmount <= 0 {
			bot.SendPM(ctx, pm.Nick, "Bet amount must be greater than 0.")
			return
		}

		// 2) Parse the choice ("odd" or "even")
		choice := strings.ToLower(tokens[2])
		if choice != "odd" && choice != "even" {
			bot.SendPM(ctx, pm.Nick, "Invalid choice. Please use 'odd' or 'even'.")
			return
		}

		// 3) Generate a random number
		randomNum := rand.Intn(100) + 1
		isRandomEven := (randomNum%2 == 0)
		userWon := false
		if (choice == "even" && isRandomEven) || (choice == "odd" && !isRandomEven) {
			userWon = true
		}

		// 4) Build result message
		resultMsg := fmt.Sprintf(
			"You bet %.8f DCR on '%s'. Random number: %d (%s).",
			betAmount.ToCoin(),
			choice,
			randomNum,
			func() string {
				if isRandomEven {
					return "even"
				}
				return "odd"
			}(),
		)

		var uid zkidentity.ShortID
		uid.FromBytes(pm.Uid)

		// 5) Pay out if the user won
		if userWon {
			payout := betAmount * 2 // double the bet for demonstration
			err := bot.PayTip(ctx, uid, payout, 3)
			if err != nil {
				fmt.Println("Error sending tip:", err)
				bot.SendPM(ctx, pm.Nick,
					resultMsg+" You won, but there was an error sending your tip: "+err.Error())
				return
			}
			bot.SendPM(ctx, pm.Nick,
				fmt.Sprintf("%s Congratulations! You won %.8f DCR!", resultMsg, payout.ToCoin()))
		} else {
			bot.SendPM(ctx, pm.Nick, resultMsg+" Sorry, you lost!")
		}

	} else {
		// Fallback or help message
		bot.SendPM(ctx, pm.Nick, "Usage: bet <amount in DCR> <odd|even>")
	}
}

func realMain() error {
	flag.Parse()

	// Expand and clean the app root path
	appRoot := utils.CleanAndExpandPath(*flagAppRoot)

	// Load bot configuration
	cfg, err := config.LoadBotConfig(appRoot, "bettingbot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create channels for handling PMs and tips
	pmChan := make(chan types.ReceivedPM)
	tipChan := make(chan types.ReceivedTip)
	tipProgressChan := make(chan types.TipProgressEvent)

	// Set up PM channel
	cfg.PMChan = pmChan

	// Set up tip channels
	cfg.TipProgressChan = tipProgressChan
	cfg.TipReceivedChan = tipChan

	// Create the bot
	bot, err := kit.NewBot(cfg)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	log := bot.LogBackend.Logger("BettingBot")
	cfg.PMLog = bot.LogBackend.Logger("PM")
	cfg.TipLog = bot.LogBackend.Logger("TIP")
	cfg.TipReceivedLog = bot.LogBackend.Logger("TIP_RECEIVED")

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received shutdown signal: %v", sig)
		bot.Close()
		cancel()
	}()

	// Handle PMs
	go func() {
		for pm := range pmChan {
			handlePM(ctx, bot, &pm)
		}
	}()

	// Handle received tips
	go func() {
		for tip := range tipChan {
			var userID zkidentity.ShortID
			userID.FromBytes(tip.Uid)

			log.Infof("Tip received: %.8f DCR from %s",
				dcrutil.Amount(tip.AmountMatoms/1e3).ToCoin(),
				userID.String())

			bot.SendPM(ctx, userID.String(),
				fmt.Sprintf("Thank you for the tip of %.8f DCR!",
					dcrutil.Amount(tip.AmountMatoms/1e3).ToCoin()))

			bot.AckTipReceived(ctx, tip.SequenceId)
		}
	}()

	// Handle tip progress updates
	go func() {
		for progress := range tipProgressChan {
			log.Infof("Tip progress event (sequence ID: %d)", progress.SequenceId)

			// Acknowledge tip progress
			err := bot.AckTipProgress(ctx, progress.SequenceId)
			if err != nil {
				log.Errorf("Failed to acknowledge tip progress: %v", err)
			}
		}
	}()

	// Run the bot
	err = bot.Run(ctx)
	log.Infof("Bot exited: %v", err)
	return err
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
