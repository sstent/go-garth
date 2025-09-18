package garmin

import (
	"time"
)

// ActivityOptions for filtering activity lists
type ActivityOptions struct {
	Limit      int
	Offset     int
	ActivityType string
	DateFrom   time.Time
	DateTo     time.Time
}

// ActivityDetail represents detailed information for an activity
type ActivityDetail struct {
	Activity // Embed garmin.Activity from pkg/garmin/types.go
	Description string `json:"description"`	// Add more fields as needed
}

// Lap represents a lap in an activity
type Lap struct {
	// Define lap fields
}

// Metric represents a metric in an activity
type Metric struct {
	// Define metric fields
}

// DownloadOptions for downloading activity data
type DownloadOptions struct {
	Format    string // "gpx", "tcx", "fit", "csv"
	Original  bool   // Download original uploaded file
	OutputDir string
	Filename  string
}
