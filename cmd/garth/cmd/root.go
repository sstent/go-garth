package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sstent/go-garth/cmd/garth/cmd/activities"
	"github.com/sstent/go-garth/cmd/garth/cmd/auth"
	garthclient "github.com/sstent/go-garth/garth/client"
)

var (
	cfgFile     string
	garthClient *garthclient.Client
)

var rootCmd = &cobra.Command{
	Use:   "garth",
	Short: "Garmin Connect CLI client",
	Long: `A comprehensive CLI client for Garmin Connect.
Access your activities, health data, and statistics from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initClient()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(auth.LoginCmd)
	rootCmd.AddCommand(auth.LogoutCmd)
	rootCmd.AddCommand(auth.StatusCmd)
	rootCmd.AddCommand(activities.ActivitiesCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.garth/config.yaml)")
	rootCmd.PersistentFlags().StringP("format", "f", "table",
		"output format (table, json, csv)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false,
		"verbose output")
	rootCmd.PersistentFlags().StringP("session", "s", "",
		"session file path")

	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("session", rootCmd.PersistentFlags().Lookup("session"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, ".garth")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}
	}
}

func initClient() {
	var err error
	domain := viper.GetString("domain")
	if domain == "" {
		domain = "garmin.com"
	}

	garthClient, err = garthclient.NewClient(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	sessionPath := viper.GetString("session")
	if sessionPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		sessionPath = filepath.Join(home, ".garth", "session.json")
	}

	if err := garthClient.LoadSession(sessionPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load session: %v\n", err)
	}
}
