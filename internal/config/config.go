package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application's configuration.
type Config struct {
	Auth struct {
		Email   string `yaml:"email"`
		Domain  string `yaml:"domain"`
		Session string `yaml:"session_file"`
	} `yaml:"auth"`

	Output struct {
		Format string `yaml:"format"`
		File   string `yaml:"file"`
	} `yaml:"output"`

	Cache struct {
		Enabled bool          `yaml:"enabled"`
		TTL     time.Duration `yaml:"ttl"`
		Dir     string        `yaml:"dir"`
	} `yaml:"cache"`
}

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Auth: struct {
			Email   string `yaml:"email"`
			Domain  string `yaml:"domain"`
			Session string `yaml:"session_file"`
		}{
			Domain:  "garmin.com",
			Session: filepath.Join(UserConfigDir(), "session.json"),
		},
		Output: struct {
			Format string `yaml:"format"`
			File   string `yaml:"file"`
		}{
			Format: "table",
		},
		Cache: struct {
			Enabled bool          `yaml:"enabled"`
			TTL     time.Duration `yaml:"ttl"`
			Dir     string        `yaml:"dir"`
		}{
			Enabled: true,
			TTL:     24 * time.Hour,
			Dir:     filepath.Join(UserCacheDir(), "cache"),
		},
	}
}

// LoadConfig loads configuration from the specified path.
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return default config if file doesn't exist
		}
		return nil, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves the configuration to the specified path.
func SaveConfig(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// InitConfig ensures the config directory and default config file exist.
func InitConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Ensure config directory exists
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	// Check if config file exists, if not, create it with default values
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := SaveConfig(path, config); err != nil {
			return nil, err
		}
	}

	return LoadConfig(path)
}

// UserConfigDir returns the user's configuration directory for garth.
func UserConfigDir() string {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "garth")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "garth")
}

// UserCacheDir returns the user's cache directory for garth.
func UserCacheDir() string {
	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		return filepath.Join(xdgCacheHome, "garth")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "garth")
}
