package internal

import (
	"encoding/json"
	"os"
	"path/filepath"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
)

type PlatformFilterMode string

const (
	PlatformFilterMatchDevice PlatformFilterMode = "match_device"
	PlatformFilterAll         PlatformFilterMode = "all"
)

type Config struct {
	PlatformFilter PlatformFilterMode `json:"platform_filter"`
}

var configInstance *Config

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			config := &Config{
				PlatformFilter: PlatformFilterMatchDevice,
			}
			configInstance = config
			return config, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Apply defaults for missing values
	if config.PlatformFilter == "" {
		config.PlatformFilter = PlatformFilterMatchDevice
	}

	configInstance = &config
	return &config, nil
}

func SaveConfig(config *Config) error {
	logger := gaba.GetLogger()

	if config.PlatformFilter == "" {
		config.PlatformFilter = PlatformFilterMatchDevice
	}

	configPath := getConfigPath()

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create config directory", "error", err)
		return err
	}

	pretty, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal config to JSON", "error", err)
		return err
	}

	if err := os.WriteFile(configPath, pretty, 0644); err != nil {
		logger.Error("Failed to write config file", "error", err)
		return err
	}

	configInstance = config
	return nil
}

func GetConfig() *Config {
	if configInstance == nil {
		config, err := LoadConfig()
		if err != nil {
			return &Config{PlatformFilter: PlatformFilterMatchDevice}
		}
		return config
	}
	return configInstance
}

func getConfigPath() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return "config.json"
	}
	return filepath.Join(utils.GetUserDataDir(), "config.json")
}
