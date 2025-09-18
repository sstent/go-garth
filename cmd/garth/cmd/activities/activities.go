package activities

import (
	"github.com/spf13/cobra"
)

var ActivitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "Manage activities from Garmin Connect",
	Long:  `Commands for listing, downloading, and managing activities from Garmin Connect`,
}

func init() {
	ActivitiesCmd.AddCommand(ListCmd)
}
