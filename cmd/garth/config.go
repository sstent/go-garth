package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"go-garth/internal/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage garth configuration",
	Long:  `Allows you to initialize, show, and manage garth's configuration file.`, 
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a default config file",
	Long:  `Creates a default garth configuration file in the standard location ($HOME/.config/garth/config.yaml) if one does not already exist.`, 
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := filepath.Join(config.UserConfigDir(), "config.yaml")
		_, err := config.InitConfig(configPath)
		if err != nil {
			return fmt.Errorf("error initializing config: %w", err)
		}
		fmt.Printf("Default config file initialized at: %s\n", configPath)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	Long:  `Displays the currently loaded garth configuration, including values from the config file and environment variables.`, 
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("error marshaling config to YAML: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}