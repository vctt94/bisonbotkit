package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vctt94/bisonbotkit/utils"
)

var (
	defaultBRDir = utils.AppDataDir("brclient", false)
)

// ClientConfig holds all configuration options for a Bison Relay client
type ClientConfig struct {
	ServerAddr      string
	RPCURL          string
	BRClientCert    string
	BRClientRPCCert string
	BRClientRPCKey  string
	GRPCServerCert  string
	RPCUser         string
	RPCPass         string
	// Logging-related fields
	LogFile        string // Path to the log file
	Debug          string // Debug level string
	MaxLogFiles    int    // Maximum number of log files to keep
	MaxBufferLines int    // Maximum number of log lines to buffer
}

// Write the configuration to a file.
func writeClientConfigFile(cfg *ClientConfig, configPath string) error {
	configData := fmt.Sprintf(
		`serveraddr=%s
rpcurl=%s
brclientcert=%s
brclientrpccert=%s
brclientrpckey=%s
grpcservercert=%s
rpcuser=%s
rpcpass=%s
logfile=%s
debug=%s
maxlogfiles=%d
maxbufferlines=%d
`,
		cfg.ServerAddr,
		cfg.RPCURL,
		cfg.BRClientCert,
		cfg.BRClientRPCCert,
		cfg.BRClientRPCKey,
		cfg.GRPCServerCert,
		cfg.RPCUser,
		cfg.RPCPass,
		cfg.LogFile,
		cfg.Debug,
		cfg.MaxLogFiles,
		cfg.MaxBufferLines,
	)

	return os.WriteFile(configPath, []byte(configData), 0600)
}

// parseClientConfigFile parses the config file at the given path into a ClientConfig struct.
func parseClientConfigFile(configPath string) (*ClientConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &ClientConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "serveraddr":
			cfg.ServerAddr = value
		case "rpcurl":
			cfg.RPCURL = value
		case "brclientcert":
			cfg.BRClientCert = value
		case "brclientrpccert":
			cfg.BRClientRPCCert = value
		case "brclientrpckey":
			cfg.BRClientRPCKey = value
		case "grpcservercert":
			cfg.GRPCServerCert = value
		case "rpcuser":
			cfg.RPCUser = value
		case "rpcpass":
			cfg.RPCPass = value
		case "logfile":
			cfg.LogFile = value
		case "debug":
			cfg.Debug = value
		case "maxlogfiles":
			fmt.Sscanf(value, "%d", &cfg.MaxLogFiles)
		case "maxbufferlines":
			fmt.Sscanf(value, "%d", &cfg.MaxBufferLines)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadClientConfig attempts to load the client config from the default locations.
func LoadClientConfig(configPath string, fileName string) (*ClientConfig, error) {
	defaultConfigPath := utils.AppDataDir(fileName, false)
	// If configPath is empty, use defaultConfigPath
	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, err
	}

	// Try to load existing config
	fullPath := filepath.Join(configPath, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return parseClientConfigFile(fullPath)
	}

	rpcUser, err := utils.GenerateRandomString(8)
	if err != nil {
		return nil, err
	}
	rpcPass, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, err
	}
	// Create default config
	cfg := &ClientConfig{
		ServerAddr:      "127.0.0.1:50051",
		RPCURL:          "wss://127.0.0.1:7676/ws",
		BRClientCert:    filepath.Join(defaultBRClientDir, "rpc.cert"),
		BRClientRPCCert: filepath.Join(defaultBRClientDir, "rpc-client.cert"),
		BRClientRPCKey:  filepath.Join(defaultBRClientDir, "rpc-client.key"),
		GRPCServerCert:  filepath.Join(configPath, "server.cert"),
		RPCUser:         rpcUser,
		RPCPass:         rpcPass,
		LogFile:         filepath.Join(configPath, "logs", fileName),
		Debug:           "info",
		MaxLogFiles:     5,
		MaxBufferLines:  1000,
	}

	// Write default config
	if err := writeClientConfigFile(cfg, fullPath); err != nil {
		return nil, err
	}

	return cfg, nil
}
