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
	DataDir string

	RPCURL          string
	BRClientCert    string
	BRClientRPCCert string
	BRClientRPCKey  string
	RPCUser         string
	RPCPass         string
	AuthMode        string
	// Logging-related fields
	LogFile        string // Path to the log file
	Debug          string // Debug level string
	MaxLogFiles    int    // Maximum number of log files to keep
	MaxBufferLines int    // Maximum number of log lines to buffer

	// Store additional config values that aren't explicitly defined
	ExtraConfig map[string]string
}

// GetString gets a configuration value as string, checking extra config first
func (c *ClientConfig) GetString(key string) string {
	switch key {
	case "brrpcurl":
		return c.RPCURL
	case "brclientcert":
		return c.BRClientCert
	case "brclientrpccert":
		return c.BRClientRPCCert
	case "brclientrpckey":
		return c.BRClientRPCKey
	case "rpcuser":
		return c.RPCUser
	case "rpcpass":
		return c.RPCPass
	case "authmode":
		return c.AuthMode
	case "logfile":
		return c.LogFile
	case "debug":
		return c.Debug
	default:
		if c.ExtraConfig != nil {
			return c.ExtraConfig[key]
		}
		return ""
	}
}

// SetString sets a configuration value as string
func (c *ClientConfig) SetString(key, value string) {
	switch key {
	case "brrpcurl":
		c.RPCURL = value
	case "brclientcert":
		c.BRClientCert = value
	case "brclientrpccert":
		c.BRClientRPCCert = value
	case "brclientrpckey":
		c.BRClientRPCKey = value
	case "rpcuser":
		c.RPCUser = value
	case "rpcpass":
		c.RPCPass = value
	case "authmode":
		c.AuthMode = value
	case "logfile":
		c.LogFile = value
	case "debug":
		c.Debug = value
	default:
		if c.ExtraConfig == nil {
			c.ExtraConfig = make(map[string]string)
		}
		c.ExtraConfig[key] = value
	}
}

// GetInt gets a configuration value as int
func (c *ClientConfig) GetInt(key string) int {
	switch key {
	case "maxlogfiles":
		return c.MaxLogFiles
	case "maxbufferlines":
		return c.MaxBufferLines
	default:
		if c.ExtraConfig != nil {
			if val, ok := c.ExtraConfig[key]; ok {
				var intVal int
				fmt.Sscanf(val, "%d", &intVal)
				return intVal
			}
		}
		return 0
	}
}

// SetInt sets a configuration value as int
func (c *ClientConfig) SetInt(key string, value int) {
	switch key {
	case "maxlogfiles":
		c.MaxLogFiles = value
	case "maxbufferlines":
		c.MaxBufferLines = value
	default:
		if c.ExtraConfig == nil {
			c.ExtraConfig = make(map[string]string)
		}
		c.ExtraConfig[key] = fmt.Sprintf("%d", value)
	}
}

// Write the configuration to a file.
func writeClientConfigFile(cfg *ClientConfig, configPath string) error {
	configData := fmt.Sprintf(
		`brrpcurl=%s
brclientcert=%s
brclientrpccert=%s
brclientrpckey=%s
rpcuser=%s
rpcpass=%s
authmode=%s
logfile=%s
debug=%s
maxlogfiles=%d
maxbufferlines=%d
`,
		cfg.RPCURL,
		cfg.BRClientCert,
		cfg.BRClientRPCCert,
		cfg.BRClientRPCKey,
		cfg.RPCUser,
		cfg.RPCPass,
		cfg.AuthMode,
		cfg.LogFile,
		cfg.Debug,
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

// parseClientConfigFile parses the config file at the given path into a ClientConfig struct.
func parseClientConfigFile(configPath string) (*ClientConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &ClientConfig{
		ExtraConfig: make(map[string]string),
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

		switch key {
		case "brrpcurl":
			cfg.RPCURL = value
		case "brclientcert":
			cfg.BRClientCert = value
		case "brclientrpccert":
			cfg.BRClientRPCCert = value
		case "brclientrpckey":
			cfg.BRClientRPCKey = value
		case "rpcuser":
			cfg.RPCUser = value
		case "rpcpass":
			cfg.RPCPass = value
		case "authmode":
			cfg.AuthMode = value
		case "logfile":
			cfg.LogFile = value
		case "debug":
			cfg.Debug = value
		case "maxlogfiles":
			fmt.Sscanf(value, "%d", &cfg.MaxLogFiles)
		case "maxbufferlines":
			fmt.Sscanf(value, "%d", &cfg.MaxBufferLines)
		default:
			cfg.ExtraConfig[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate required fields are present and accessible
	var missing []string
	if cfg.RPCURL == "" {
		missing = append(missing, "brrpcurl")
	}
	if cfg.BRClientCert == "" {
		missing = append(missing, "brclientcert")
	}
	if cfg.BRClientRPCCert == "" {
		missing = append(missing, "brclientrpccert")
	}
	if cfg.BRClientRPCKey == "" {
		missing = append(missing, "brclientrpckey")
	}
	if cfg.RPCUser == "" {
		missing = append(missing, "rpcuser")
	}
	if cfg.RPCPass == "" {
		missing = append(missing, "rpcpass")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required fields in client config: %s", strings.Join(missing, ", "))
	}

	// Ensure cert/key files exist so we fail fast with a clear error
	for _, fp := range []struct{ name, path string }{
		{"brclientcert", cfg.BRClientCert},
		{"brclientrpccert", cfg.BRClientRPCCert},
		{"brclientrpckey", cfg.BRClientRPCKey},
	} {
		if _, err := os.Stat(fp.path); err != nil {
			return nil, fmt.Errorf("%s file not accessible at %q: %v", fp.name, fp.path, err)
		}
	}

	return cfg, nil
}

// LoadClientConfig attempts to load the client config (.conf) from the default locations.
func LoadClientConfig(configPath string, fileName string) (*ClientConfig, error) {
	// Check if fileName has .conf extension
	if !strings.HasSuffix(fileName, ".conf") {
		return nil, fmt.Errorf("filename must have .conf extension, got: %s", fileName)
	}

	// Get app name by removing .conf extension
	appName := strings.TrimSuffix(fileName, ".conf")

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

	// Create default config
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
		RPCURL:          "wss://127.0.0.1:7676/ws",
		BRClientCert:    filepath.Join(defaultBRDir, "rpc.cert"),
		BRClientRPCCert: filepath.Join(defaultBRDir, "rpc-client.cert"),
		BRClientRPCKey:  filepath.Join(defaultBRDir, "rpc-client.key"),
		RPCUser:         rpcUser,
		RPCPass:         rpcPass,
		AuthMode:        "basic",
		LogFile:         filepath.Join(configPath, "logs", appName+".log"),
		Debug:           "info",
		MaxLogFiles:     5,
		MaxBufferLines:  1000,
		ExtraConfig:     make(map[string]string),
	}

	// Write default config
	if err := writeClientConfigFile(cfg, fullPath); err != nil {
		return nil, err
	}

	return cfg, nil
}
