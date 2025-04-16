# BisonBotKit

BisonBotKit is a Go library for creating and managing bots for the BisonRelay messaging platform. The library provides a simplified interface for connecting to BisonRelay services, and handling configurations.

## Usage

### Creating a Bot with Configuration and Logging

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

func main() {
	// Get application data directory
	appdata := utils.AppDataDir("mybot", false)
	
	// Load bot configuration
	cfg, err := config.LoadBotConfig(appdata, "mybot.conf")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up logging
	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appdata, "logs", "mybot.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
	})
	if err != nil {
		fmt.Printf("Failed to create log backend: %v\n", err)
		os.Exit(1)
	}
	defer logBackend.Close()

	// Get a logger for your application
	log := logBackend.Logger("Bot")

	// Create channels for handling messages (if needed)
	pmChan := make(chan types.ReceivedPM)
	// Assign the channel to the config
	cfg.PMChan = pmChan
	cfg.PMLog = logBackend.Logger("PM")

	// Create and run the bot
	bot, err := bisonbotkit.NewBot(cfg, logBackend)
	if err != nil {
		log.Errorf("Failed to create bot: %v", err)
		os.Exit(1)
	}

	// Add a goroutine to handle incoming private messages
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg)
			// Handle the message here
		}
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Infof("Received shutdown signal")
		bot.Close()
		cancel()
	}()

	// Run the bot
	if err := bot.Run(ctx); err != nil {
		log.Errorf("Bot error: %v", err)
		os.Exit(1)
	}
}
```

## Configuration

The library supports configuration through both configuration files and command-line flags. Configuration can be loaded from:

- Specified path (if provided)
- Application data directory (`~/.mybot/mybot.conf`)
- Current directory (`./mybot.conf`)

Configuration supports various settings including:

### Bot Configuration Example
```
# Data directory for storing bot data
datadir=/path/to/data/dir

# RPC connection settings
rpcurl=wss://127.0.0.1:7676/ws
grpchost=127.0.0.1
grpcport=50051
httpport=8888

# Certificate paths for secure connections
servercertpath=/path/to/rpc.cert
clientcertpath=/path/to/rpc-client.cert
clientkeypath=/path/to/rpc-client.key

# RPC authentication
rpcuser=your_rpc_username
rpcpass=your_rpc_password

# Bot-specific settings
isf2p=false              # Whether the bot is for F2P (Free-to-Play) mode
minbetamt=0.00000001     # Minimum bet amount

# Logging level (debug, info, warn, error)
debug=debug
```

### Client Configuration Example
```
# Server connection settings
serveraddr=127.0.0.1:50051
rpcurl=wss://localhost:7777/ws

# Certificate paths for secure connections
servercertpath=/path/to/rpc.cert
clientcertpath=/path/to/rpc-client.cert
clientkeypath=/path/to/rpc-client.key
grpcservercert=/path/to/server.cert

# RPC authentication
rpcuser=your_rpc_username
rpcpass=your_rpc_password
```

Each setting is explained below:

#### Bot Configuration Settings
- `datadir`: Directory where the bot stores its data
- `rpcurl`: WebSocket URL for RPC connection
- `grpchost`: Host address for gRPC connection
- `grpcport`: Port for gRPC connection
- `httpport`: Port for HTTP server
- `servercertpath`: Path to server certificate for secure connections
- `clientcertpath`: Path to client certificate
- `clientkeypath`: Path to client private key
- `rpcuser`: Username for RPC authentication
- `rpcpass`: Password for RPC authentication
- `isf2p`: Boolean flag for F2P mode
- `minbetamt`: Minimum bet amount for the bot
- `debug`: Logging level (debug, info, warn, error)

#### Client Configuration Settings
- `serveraddr`: Server address in host:port format
- `rpcurl`: WebSocket URL for RPC connection
- `servercertpath`: Path to server certificate for secure connections
- `clientcertpath`: Path to client certificate
- `clientkeypath`: Path to client private key
- `grpcservercert`: Path to gRPC server certificate
- `rpcuser`: Username for RPC authentication
- `rpcpass`: Password for RPC authentication

## Logging Features

BisonBotKit includes a powerful logging system with the following features:

- **Log Rotation**: Automatically rotates log files to keep disk usage under control
- **Log Levels**: Different verbosity levels for different subsystems
- **In-Memory Buffer**: Keeps recent log messages in memory for quick access
- **Subsystem Loggers**: Creates different loggers for different parts of your application

## License

bisonbotkit is licensed under the [copyfree](http://copyfree.org) ISC License.

## Contributors

- [dajohi](https://github.com/dajohi) 
- [vctt94](https://github.com/vctt94)
