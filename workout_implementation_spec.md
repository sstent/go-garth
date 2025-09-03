# Workout Service Implementation Specification

## Overview
This document provides the technical specification for implementing the complete WorkoutService functionality in the go-garth library, following the established patterns from the ActivityService.

## Data Structures

### Workout (Summary)
```go
type Workout struct {
    WorkoutID   int64     `json:"workoutId"`
    Name        string    `json:"workoutName"`
    Type        string    `json:"workoutType"`
    Description string    `json:"description"`
    CreatedDate time.Time `json:"createdDate"`
    UpdatedDate time.Time `json:"updatedDate"`
    OwnerID     int64     `json:"ownerId"`
    IsPublic    bool      `json:"isPublic"`
    SportType   string    `json:"sportType"`
    SubSportType string   `json:"subSportType"`
}
```

### WorkoutDetails (Full)
```go
type WorkoutDetails struct {
    WorkoutID       int64           `json:"workoutId"`
    Name            string          `json:"workoutName"`
    Description     string          `json:"description"`
    Type            string          `json:"workoutType"`
    CreatedDate     time.Time       `json:"createdDate"`
    UpdatedDate     time.Time       `json:"updatedDate"`
    OwnerID         int64           `json:"ownerId"`
    IsPublic        bool            `json:"isPublic"`
    SportType       string          `json:"sportType"`
    SubSportType    string          `json:"subSportType"`
    WorkoutSegments []WorkoutSegment `json:"workoutSegments"`
    EstimatedDuration int           `json:"estimatedDuration"`
    EstimatedDistance float64       `json:"estimatedDistance"`
    TrainingLoad    float64         `json:"trainingLoad"`
    Tags            []string        `json:"tags"`
}
```

### WorkoutSegment
```go
type WorkoutSegment struct {
    SegmentID   int64           `json:"segmentId"`
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Order       int             `json:"order"`
    Exercises   []WorkoutExercise `json:"exercises"`
}
```

### WorkoutExercise
```go
type WorkoutExercise struct {
    ExerciseID   int64  `json:"exerciseId"`
    Name         string `json:"name"`
    Category     string `json:"category"`
    Type         string `json:"type"`
    Duration     int    `json:"duration,omitempty"`
    Distance     float64 `json:"distance,omitempty"`
    Repetitions  int    `json:"repetitions,omitempty"`
    Weight       float64 `json:"weight,omitempty"`
    RestInterval int    `json:"restInterval,omitempty"`
}
```

### WorkoutListOptions
```go
type WorkoutListOptions struct {
    Limit        int
    Offset       int
    StartDate    time.Time
    EndDate      time.Time
    WorkoutType  string
    Type         string
    Status       string
    SportType    string
    NameContains string
    OwnerID      int64
    IsPublic     *bool
    SortBy       string
    SortOrder    string
}
```

### WorkoutUpdate
```go
type WorkoutUpdate struct {
    Name        string    `json:"workoutName,omitempty"`
    Description string    `json:"description,omitempty"`
    Type        string    `json:"workoutType,omitempty"`
    SportType   string    `json:"sportType,omitempty"`
    IsPublic    *bool     `json:"isPublic,omitempty"`
    Tags        []string  `json:"tags,omitempty"`
}
```

## API Endpoints

### Base URL Pattern
All workout endpoints follow the pattern: `/workout-service/workout`

### Specific Endpoints
- **GET** `/workout-service/workout` - List workouts (with query parameters)
- **GET** `/workout-service/workout/{workoutId}` - Get workout details
- **POST** `/workout-service/workout` - Create new workout
- **PUT** `/workout-service/workout/{workoutId}` - Update existing workout
- **DELETE** `/workout-service/workout/{workoutId}` - Delete workout
- **GET** `/download-service/export/{format}/workout/{workoutId}` - Export workout

## Method Signatures

### WorkoutService Methods
```go
func (s *WorkoutService) List(ctx context.Context, opts WorkoutListOptions) ([]Workout, error)
func (s *WorkoutService) Get(ctx context.Context, workoutID int64) (*WorkoutDetails, error)
func (s *WorkoutService) Create(ctx context.Context, workout WorkoutDetails) (*WorkoutDetails, error)
func (s *WorkoutService) Update(ctx context.Context, workoutID int64, update WorkoutUpdate) (*WorkoutDetails, error)
func (s *WorkoutService) Delete(ctx context.Context, workoutID int64) error
func (s *WorkoutService) Export(ctx context.Context, workoutID int64, format string) (io.ReadCloser, error)
```

## Testing Strategy

### Test Coverage Requirements
- All public methods must have unit tests
- Test both success and error scenarios
- Use httptest for HTTP mocking
- Test JSON marshaling/unmarshaling
- Test query parameter construction

### Test File Structure
- File: `workouts_test.go`
- Test functions:
  - `TestWorkoutService_List`
  - `TestWorkoutService_Get`
  - `TestWorkoutService_Create`
  - `TestWorkoutService_Update`
  - `TestWorkoutService_Delete`
  - `TestWorkoutService_Export`

## Error Handling
- Follow the same pattern as ActivityService
- Use APIError for HTTP-related errors
- Include proper status codes and messages
- Handle JSON parsing errors appropriately

## Implementation Notes
- Use the same HTTP client pattern as ActivityService
- Follow consistent naming conventions
- Ensure proper context handling
- Include appropriate defer statements for resource cleanup
- Use strconv.FormatInt for ID conversion in URLs