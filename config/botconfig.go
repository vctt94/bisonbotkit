package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/slog"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/vctt94/bisonbotkit/utils"
)

var (
	defaultBRClientDir = utils.AppDataDir("brclient", false)
)

// BotConfig holds all configuration options for a Bison Relay bot
type BotConfig struct {
	DataDir string

	RPCURL         string
	ServerCertPath string
	ClientCertPath string
	ClientKeyPath  string

	GCChan     chan<- types.GCReceivedMsg
	GCLog      slog.Logger
	InviteChan chan<- types.ReceivedGCInvite

	PMChan chan<- types.ReceivedPM
	PMLog  slog.Logger

	PostChan chan<- types.ReceivedPost
	PostLog  slog.Logger

	PostStatusChan chan<- types.ReceivedPostStatus
	PostStatusLog  slog.Logger

	TipProgressChan chan<- types.TipProgressEvent
	TipLog          slog.Logger

	TipReceivedChan chan<- types.ReceivedTip
	TipReceivedLog  slog.Logger

	KXChan chan<- types.KXCompleted
	KXLog  slog.Logger

	RPCUser string
	RPCPass string
	Debug   string
	// Logging-related fields
	LogFile        string // Path to the log file
	MaxLogFiles    int    // Maximum number of log files to keep
	MaxBufferLines int    // Maximum number of log lines to buffer

	// Store additional config values that aren't explicitly defined
	ExtraConfig map[string]string
}

// Write the configuration to a file.
func writeConfigFile(cfg *BotConfig, configPath string) error {
	// Build the basic config string with known fields
	configData := fmt.Sprintf(
		`datadir=%s
brrpcurl=%s
servercertpath=%s
clientcertpath=%s
clientkeypath=%s
rpcuser=%s
rpcpass=%s
debug=%s
logfile=%s
maxlogfiles=%d
maxbufferlines=%d
`,
		cfg.DataDir,
		cfg.RPCURL,
		cfg.ServerCertPath,
		cfg.ClientCertPath,
		cfg.ClientKeyPath,
		cfg.RPCUser,
		cfg.RPCPass,
		cfg.Debug,
		cfg.LogFile,
		cfg.MaxLogFiles,
		cfg.MaxBufferLines,
	)

	// Add any extra config fields
	var extraConfig strings.Builder
	for key, value := range cfg.ExtraConfig {
		extraConfig.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	// Combine all config data
	fullConfig := configData + extraConfig.String()

	return os.WriteFile(configPath, []byte(fullConfig), 0600)
}

// parseConfigFile parses the config file at the given path into a BotConfig struct.
func parseConfigFile(configPath string) (*BotConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &BotConfig{
		ExtraConfig: make(map[string]string), // Initialize the map
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Process known config fields
		handled := true
		switch key {
		case "datadir":
			cfg.DataDir = value
		case "brrpcurl":
			cfg.RPCURL = value
		case "servercertpath":
			cfg.ServerCertPath = value
		case "clientcertpath":
			cfg.ClientCertPath = value
		case "clientkeypath":
			cfg.ClientKeyPath = value
		case "rpcuser":
			cfg.RPCUser = value
		case "rpcpass":
			cfg.RPCPass = value
		case "debug":
			cfg.Debug = value
		case "logfile":
			cfg.LogFile = value
		case "maxlogfiles":
			fmt.Sscanf(value, "%d", &cfg.MaxLogFiles)
		case "maxbufferlines":
			fmt.Sscanf(value, "%d", &cfg.MaxBufferLines)
		default:
			handled = false
		}

		// If this is not a known field, store it in the ExtraConfig map
		if !handled {
			cfg.ExtraConfig[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadBotConfig attempts to load the bot config from the default locations.
func LoadBotConfig(configPath string, fileName string) (*BotConfig, error) {
	// Check if fileName has .conf extension
	if !strings.HasSuffix(fileName, ".conf") {
		return nil, fmt.Errorf("filename must have .conf extension, got: %s", fileName)
	}

	// Get app name by removing .conf extension
	appName := strings.TrimSuffix(fileName, ".conf")

	defaultConfigPath := utils.AppDataDir(fileName, false)
	configPath = utils.CleanAndExpandPath(configPath)
	// If configPath is empty, use defaultConfigPath
	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(configPath, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		cfg, err := parseConfigFile(fullPath)
		if err == nil {
			return cfg, nil
		}
	}

	// If we get here, either the file doesn't exist or couldn't be parsed
	// Generate new credentials and create default config
	rpcUser, err := utils.GenerateRandomString(8)
	if err != nil {
		return nil, err
	}
	rpcPass, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	cfg := &BotConfig{
		DataDir:        configPath,
		RPCURL:         "wss://127.0.0.1:7676/ws",
		ServerCertPath: filepath.Join(defaultBRClientDir, "rpc.cert"),
		ClientCertPath: filepath.Join(defaultBRClientDir, "rpc-client.cert"),
		ClientKeyPath:  filepath.Join(defaultBRClientDir, "rpc-client.key"),
		RPCUser:        rpcUser,
		RPCPass:        rpcPass,
		Debug:          "info",
		LogFile:        filepath.Join(configPath, "logs", appName+".log"),
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
		ExtraConfig:    make(map[string]string), // Initialize the map for new configs
	}

	// Write default config
	if err := writeConfigFile(cfg, fullPath); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}

	return cfg, nil
}
