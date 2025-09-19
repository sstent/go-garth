package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"go-garth/pkg/garmin"

	"github.com/spf13/cobra"
)

var (
	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authentication management",
		Long:  `Manage authentication with Garmin Connect, including login, logout, and status.`,
	}

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to Garmin Connect",
		Long:  `Login to Garmin Connect interactively or using provided credentials.`,
		RunE:  runLogin,
	}

	logoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Logout from Garmin Connect",
		Long:  `Clear the current Garmin Connect session.`,
		RunE:  runLogout,
	}

	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show Garmin Connect authentication status",
		Long:  `Display the current authentication status and session information.`,
		RunE:  runStatus,
	}

	refreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh Garmin Connect session tokens",
		Long:  `Refresh the authentication tokens for the current Garmin Connect session.`,
		RunE:  runRefresh,
	}

	loginEmail        string
	loginPassword     string
	passwordStdinFlag bool
)

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email for Garmin Connect login")
	loginCmd.Flags().BoolVarP(&passwordStdinFlag, "password-stdin", "p", false, "Read password from stdin")

	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(refreshCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	var email, password string
	var err error

	if loginEmail != "" {
		email = loginEmail
	} else {
		fmt.Print("Enter Garmin Connect email: ")
		_, err = fmt.Scanln(&email)
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
	}

	if passwordStdinFlag {
		fmt.Print("Enter password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password from stdin: %w", err)
		}
		password = string(passwordBytes)
		fmt.Println() // Newline after password input
	} else {
		fmt.Print("Enter password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
		fmt.Println() // Newline after password input
	}

	// Create client
	// TODO: Domain should be configurable
	garminClient, err := garmin.NewClient("www.garmin.com")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Try to load existing session first
	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		fmt.Println("No existing session found or session invalid, logging in with credentials...")

		if err := garminClient.Login(email, password); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Save session for future use
		if err := garminClient.SaveSession(sessionFile); err != nil {
			fmt.Printf("Failed to save session: %v\n", err)
		}
	} else {
		fmt.Println("Loaded existing session")
	}

	fmt.Println("Login successful!")
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	sessionFile := "garmin_session.json" // TODO: Make session file configurable

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		fmt.Println("No active session to log out from.")
		return nil
	}

	if err := os.Remove(sessionFile); err != nil {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	fmt.Println("Logged out successfully. Session cleared.")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	sessionFile := "garmin_session.json" // TODO: Make session file configurable

	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(sessionFile); err != nil {
		fmt.Println("Not logged in or session expired.")
		return nil
	}

	fmt.Println("Logged in. Session is active.")
	// TODO: Add more detailed status information, e.g., session expiry
	return nil
}

func runRefresh(cmd *cobra.Command, args []string) error {
	sessionFile := "garmin_session.json" // TODO: Make session file configurable

	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("cannot refresh: no active session found: %w", err)
	}

	fmt.Println("Attempting to refresh session...")
	if err := garminClient.RefreshSession(); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	if err := garminClient.SaveSession(sessionFile); err != nil {
		fmt.Printf("Failed to save refreshed session: %v\n", err)
	}

	fmt.Println("Session refreshed successfully.")
	return nil
}
