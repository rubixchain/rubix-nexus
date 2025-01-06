package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// LoadConfig loads the configuration from the config file
func LoadConfig(homeDir string) (*Config, error) {
	configPath := filepath.Join(homeDir, ".rubix-nexus", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GenerateConfig creates a default configuration file
func GenerateConfig(homeDir string) error {
	configDir := filepath.Join(homeDir, ".rubix-nexus")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	defaultConfig := Config{
		Network: NetworkConfig{
			DeployerNodeURL: "http://localhost:20011",
		},
	}

	configContent, err := toml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig checks if the configuration file is valid
func ValidateConfig(homeDir string) error {
	configPath := filepath.Join(homeDir, ".rubix-nexus", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found at %s", configPath)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("invalid config file format: %w", err)
	}

	// Validate required fields
	if config.Network.DeployerNodeURL == "" {
		return fmt.Errorf("deployer_node_url is required")
	}

	return nil
}
