// Package globalconfig manages the global Loom configuration file.
// This file stores information about configured thread stores.
package globalconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	// ConfigDirName is the name of the directory where Loom stores its global config.
	ConfigDirName = "loom"
	// ConfigFileName is the name of the global Loom configuration file.
	ConfigFileName = "loom.yaml"
)

// Store represents a configured thread store.
type Store struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // e.g., "local", "github"
	Path string `yaml:"path"` // For local type, this is the filesystem path. For github, a base URL.
}

// GlobalLoomConfig represents the structure of the global Loom configuration file.
type GlobalLoomConfig struct {
	Version string  `yaml:"version"`
	Stores  []Store `yaml:"stores,omitempty"`
}

// GetGlobalConfigPath returns the absolute path to the global Loom configuration file.
// It ensures the configuration directory exists. If LOOM_GLOBAL_DIR environment variable
// is set, it will use that as the directory containing the config file.
func GetGlobalConfigPath() (string, error) {
	var configPath string

	// Check if LOOM_GLOBAL_DIR environment variable is set
	if envDir := os.Getenv("LOOM_GLOBAL_DIR"); envDir != "" {
		configPath = envDir
	} else {
		switch runtime.GOOS {
		case "windows":
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			configPath = filepath.Join(userHomeDir, ".config", "loom")
		default: // Linux, macOS, etc.
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user config directory: %w", err)
			}
			configPath = filepath.Join(userConfigDir, "loom")
		}
	}

	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create loom config directory at %s: %w", configPath, err)
	}
	return filepath.Join(configPath, ConfigFileName), nil
}

// LoadGlobalConfig loads the global Loom configuration from the default path.
// If the file doesn't exist, it returns an empty GlobalLoomConfig with version 1.
func LoadGlobalConfig() (*GlobalLoomConfig, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	var config GlobalLoomConfig
	configData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return a new config
			return &GlobalLoomConfig{Version: "1", Stores: []Store{}}, nil
		}
		return nil, fmt.Errorf("failed to read global config file %s: %w", configPath, err)
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse global config file %s: %w", configPath, err)
	}
	if config.Stores == nil { // Ensure Stores is initialized if it was null in the YAML
		config.Stores = []Store{}
	}
	return &config, nil
}

// SaveGlobalConfig saves the global Loom configuration to the default path.
func SaveGlobalConfig(config *GlobalLoomConfig) error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	return os.WriteFile(configPath, updatedData, 0600) // 0600 for user read/write only
}
