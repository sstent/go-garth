package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sstent/go-garth"
)

func main() {
	// Create a new client (placeholder - actual client creation depends on package structure)
	// client := garth.NewClient(nil)

	// For demonstration, we'll use a mock server or skip authentication
	// In real usage, you would authenticate first:
	// err := client.Authenticate("username", "password")
	// if err != nil {
	//     log.Fatalf("Authentication failed: %v", err)
	// }

	// Example usage of the workout service
	fmt.Println("Garmin Connect Workout Service Examples")
	fmt.Println("=======================================")

	// Create a new workout
	newWorkout := garth.Workout{
		Name:        "Morning Run",
		Description: "5K easy run",
		Type:        "running",
	}
	_ = newWorkout // Prevent unused variable error

	// In a real scenario, you would do:
	// createdWorkout, err := client.Workouts.Create(context.Background(), newWorkout)
	// if err != nil {
	//     log.Printf("Failed to create workout: %v", err)
	// } else {
	//     fmt.Printf("Created workout: %+v\n", createdWorkout)
	// }

	// List workouts with options
	opts := garth.WorkoutListOptions{
		Limit:     10,
		Offset:    0,
		StartDate: time.Now().AddDate(0, -1, 0), // Last month
		EndDate:   time.Now(),
		SortBy:    "createdDate",
		SortOrder: "desc",
	}

	fmt.Printf("Workout list options: %+v\n", opts)

	// List workouts with pagination
	paginatedOpts := garth.WorkoutListOptions{
		Limit:  5,
		Offset: 10,
		SortBy: "name",
	}
	fmt.Printf("Paginated workout options: %+v\n", paginatedOpts)

	// List workouts with type and status filters
	filteredOpts := garth.WorkoutListOptions{
		Type:   "running",
		Status: "active",
		Limit:  20,
	}
	fmt.Printf("Filtered workout options: %+v\n", filteredOpts)

	// Get workout details
	workoutID := "12345"
	fmt.Printf("Would fetch workout details for ID: %s\n", workoutID)
	// workout, err := client.Workouts.Get(context.Background(), workoutID)

	// Export workout
	fmt.Printf("Would export workout %s in FIT format\n", workoutID)
	// reader, err := client.Workouts.Export(context.Background(), workoutID, "fit")

	// Search workouts
	searchOpts := garth.WorkoutListOptions{
		Limit: 5,
	}
	fmt.Printf("Would search workouts with: %+v\n", searchOpts)
	// results, err := client.Workouts.List(context.Background(), searchOpts)

	// Get workout templates
	fmt.Println("Would fetch workout templates")
	// templates, err := client.Workouts.GetWorkoutTemplates(context.Background())

	// Copy workout
	newName := "Copied Workout"
	fmt.Printf("Would copy workout %s as %s\n", workoutID, newName)
	// copied, err := client.Workouts.CopyWorkout(context.Background(), workoutID, newName)

	// Update workout
	update := garth.Workout{
		Name:        "Updated Morning Run",
		Description: "Updated description",
	}
	fmt.Printf("Would update workout %s with: %+v\n", workoutID, update)
	// updated, err := client.Workouts.Update(context.Background(), workoutID, update)

	// Delete workout
	fmt.Printf("Would delete workout: %s\n", workoutID)
	// err = client.Workouts.Delete(context.Background(), workoutID)

	fmt.Println("\nExample completed. To use with real data:")
	fmt.Println("1. Set GARMIN_USERNAME and GARMIN_PASSWORD environment variables")
	fmt.Println("2. Uncomment the authentication and API calls above")
	fmt.Println("3. Run: go run examples/workouts_example.go")
}

func init() {
	// Check if credentials are provided
	if os.Getenv("GARMIN_USERNAME") == "" || os.Getenv("GARMIN_PASSWORD") == "" {
		fmt.Println("Note: Set GARMIN_USERNAME and GARMIN_PASSWORD environment variables for real API usage")
	}
}
