# Garmin Connect Activity Examples

This directory contains examples demonstrating how to use the go-garth library to interact with Garmin Connect activities.

## Prerequisites

1. A Garmin Connect account
2. Your Garmin Connect username and password
3. Go 1.22 or later

## Setup

1. Create a `.env` file in this directory with your credentials:
   ```
   GARMIN_USERNAME=your_username
   GARMIN_PASSWORD=your_password
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Examples

### Basic Activity Listing (`activities_example.go`)

This example demonstrates basic authentication and activity listing functionality:

- Authenticates with Garmin Connect
- Lists recent activities
- Shows basic activity information

Run with:
```bash
go run activities_example.go
```

### Enhanced Activity Listing (`enhanced_example.go`)

This example provides more comprehensive activity querying capabilities:

- Lists recent activities
- Filters activities by date range
- Filters activities by activity type (e.g., running, cycling)
- Searches activities by name
- Retrieves detailed activity information including:
  - Elevation data
  - Heart rate metrics
  - Speed metrics
  - Step counts

Run with:
```bash
go run enhanced_example.go
```

## Activity Types

Common activity types you can filter by:
- `running`
- `cycling`
- `walking`
- `swimming`
- `strength_training`
- `hiking`
- `yoga`

## Output Examples

### Basic Listing
```
=== RECENT ACTIVITIES ===
Recent Activities (5 total):
==================================================
1. Morning Run [running]
   Date: 2024-01-15 07:30
   Distance: 5.20 km
   Duration: 28m
   Calories: 320

2. Evening Ride [cycling]
   Date: 2024-01-14 18:00
   Distance: 15.50 km
   Duration: 45m
   Calories: 280
```

### Detailed Activity Information
```
=== DETAILS FOR ACTIVITY: Morning Run ===
Activity ID: 123456789
Name: Morning Run
Type: running
Description: Easy morning run in the park
Start Time: 2024-01-15 07:30:00
Distance: 5.20 km
Duration: 28m
Calories: 320
Elevation Gain: 45 m
Elevation Loss: 42 m
Max Heart Rate: 165 bpm
Avg Heart Rate: 142 bpm
Max Speed: 12.5 km/h
Avg Speed: 11.1 km/h
Steps: 5200
```

## Security Notes

- Never commit your `.env` file to version control
- The `.env` file is already included in `.gitignore`
- Consider using environment variables directly in production instead of `.env` files