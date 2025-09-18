package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go-garth/internal/config"
)

var (
	cfgFile      string
	userConfigDir string
	cfg          *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "garth",
	Short: "Garmin Connect CLI tool",
	Long: `A comprehensive CLI tool for interacting with Garmin Connect.

Garth allows you to fetch your Garmin Connect data, including activities,
health stats, and more, directly from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config is loaded before any command runs
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/garth/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&userConfigDir, "config-dir", "", "config directory (default is $HOME/.config/garth)")

	rootCmd.PersistentFlags().String("output", "table", "output format (json, table, csv)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().String("date-from", "", "start date for data fetching (YYYY-MM-DD)")
	rootCmd.PersistentFlags().String("date-to", "", "end date for data fetching (YYYY-MM-DD)")

	// Bind flags to viper
	_ = viper.BindPFlag("output.format", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("dateFrom", rootCmd.PersistentFlags().Lookup("date-from"))
	_ = viper.BindPFlag("dateTo", rootCmd.PersistentFlags().Lookup("date-to"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if userConfigDir == "" {
		userConfigDir = config.UserConfigDir()
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in user's config directory with name "config" (without extension).
		viper.AddConfigPath(userConfigDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		// If config file not found, try to initialize a default one
		defaultConfigPath := filepath.Join(userConfigDir, "config.yaml")
		if _, statErr := os.Stat(defaultConfigPath); os.IsNotExist(statErr) {
			fmt.Fprintln(os.Stderr, "No config file found. Initializing default config at:", defaultConfigPath)
			var initErr error
			cfg, initErr = config.InitConfig(defaultConfigPath)
			if initErr != nil {
				fmt.Fprintln(os.Stderr, "Error initializing default config:", initErr)
				os.Exit(1)
			}
		} else if statErr != nil {
			fmt.Fprintln(os.Stderr, "Error checking for config file:", statErr)
			os.Exit(1)
		}
	}

	// Unmarshal config into our struct
	if cfg == nil { // Only unmarshal if not already initialized by InitConfig
		cfg = config.DefaultConfig() // Start with defaults
		if err := viper.Unmarshal(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "Error unmarshaling config:", err)
			os.Exit(1)
		}
	}

	// Override config with flag values
	if rootCmd.PersistentFlags().Lookup("output").Changed {
		cfg.Output.Format = viper.GetString("output.format")
	}
	// Add other flag overrides as needed
}
