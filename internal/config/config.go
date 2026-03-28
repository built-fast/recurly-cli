package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Dir returns the configuration directory path.
// Uses $XDG_CONFIG_HOME/recurly if XDG_CONFIG_HOME is set,
// otherwise ~/.config/recurly.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "recurly")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "recurly")
}

// FilePath returns the full path to the configuration file.
func FilePath() string {
	return filepath.Join(Dir(), "config.toml")
}

// Init configures Viper to read the config file. Call once at CLI startup.
func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(Dir())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("reading config: %w", err)
	}
	return nil
}

// Write persists a key-value pair to the config file.
// Creates the directory (0700) and file (0600) if needed.
func Write(key, value string) error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	path := FilePath()

	// Use a separate viper instance to avoid writing flag defaults
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(path)
	_ = v.ReadInConfig()

	v.Set(key, value)

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return os.Chmod(path, 0600)
}
