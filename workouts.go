package garth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// WorkoutService provides methods for interacting with Garmin workout data.
type WorkoutService struct {
	client *APIClient
}

// NewWorkoutService creates a new WorkoutService instance.
// client: The authenticated APIClient used to make requests.
func NewWorkoutService(client *APIClient) *WorkoutService {
	return &WorkoutService{client: client}
}

// Workout represents a Garmin workout with basic information
type Workout struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Type            string           `json:"type"`
	SportType       string           `json:"sportType"`
	SubSportType    string           `json:"subSportType"`
	CreatedDate     time.Time        `json:"createdDate"`
	UpdatedDate     time.Time        `json:"updatedDate"`
	OwnerID         int64            `json:"ownerId"`
	WorkoutSegments []WorkoutSegment `json:"workoutSegments,omitempty"`
}

// WorkoutDetails contains detailed information about a workout
type WorkoutDetails struct {
	Workout
	EstimatedDuration   int64            `json:"estimatedDuration"`
	EstimatedDistance   float64          `json:"estimatedDistance"`
	TrainingStressScore float64          `json:"trainingStressScore"`
	IntensityFactor     float64          `json:"intensityFactor"`
	WorkoutProvider     string           `json:"workoutProvider"`
	WorkoutSource       string           `json:"workoutSource"`
	WorkoutMetrics      json.RawMessage  `json:"workoutMetrics"`
	WorkoutGoals        json.RawMessage  `json:"workoutGoals"`
	WorkoutTags         []string         `json:"workoutTags"`
	WorkoutSegments     []WorkoutSegment `json:"workoutSegments"`
}

// WorkoutSegment represents a segment within a workout
type WorkoutSegment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Order       int               `json:"order"`
	Duration    int64             `json:"duration"`
	Distance    float64           `json:"distance"`
	Exercises   []WorkoutExercise `json:"exercises"`
}

// WorkoutExercise represents an exercise within a workout segment
type WorkoutExercise struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Order           int             `json:"order"`
	Duration        int64           `json:"duration"`
	Distance        float64         `json:"distance"`
	Repetitions     int             `json:"repetitions"`
	Weight          float64         `json:"weight"`
	Intensity       string          `json:"intensity"`
	ExerciseMetrics json.RawMessage `json:"exerciseMetrics"`
}

// WorkoutListOptions provides filtering options for listing workouts
type WorkoutListOptions struct {
	Limit        int
	StartDate    time.Time
	EndDate      time.Time
	SportType    string
	NameContains string
	OwnerID      int64
	Offset       int
	SortBy       string
	SortOrder    string
	Type         string
	Status       string
}

// WorkoutUpdate represents fields that can be updated on a workout
type WorkoutUpdate struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type,omitempty"`
	SportType    string `json:"sportType,omitempty"`
	SubSportType string `json:"subSportType,omitempty"`
}

// List retrieves a list of workouts for the current user with optional filters
func (s *WorkoutService) List(ctx context.Context, opts WorkoutListOptions) ([]Workout, error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if !opts.StartDate.IsZero() {
		params.Set("startDate", opts.StartDate.Format(time.RFC3339))
	}
	if !opts.EndDate.IsZero() {
		params.Set("endDate", opts.EndDate.Format(time.RFC3339))
	}
	if opts.SportType != "" {
		params.Set("sportType", opts.SportType)
	}
	if opts.Type != "" {
		params.Set("type", opts.Type)
	}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.NameContains != "" {
		params.Set("nameContains", opts.NameContains)
	}
	if opts.OwnerID > 0 {
		params.Set("ownerId", strconv.FormatInt(opts.OwnerID, 10))
	}
	if opts.SortBy != "" {
		params.Set("sortBy", opts.SortBy)
	}
	if opts.SortOrder != "" {
		params.Set("sortOrder", opts.SortOrder)
	}

	path := "/workout-service/workouts"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get workouts list",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read workouts response",
			Cause:      err,
		}
	}

	var workouts []Workout
	if err := json.Unmarshal(body, &workouts); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse workouts data",
			Cause:      err,
		}
	}

	return workouts, nil
}

// Get retrieves detailed information about a specific workout
func (s *WorkoutService) Get(ctx context.Context, id string) (*WorkoutDetails, error) {
	path := "/workout-service/workout/" + id

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get workout details",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read workout response",
			Cause:      err,
		}
	}

	var details WorkoutDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse workout data",
			Cause:      err,
		}
	}

	return &details, nil
}

// Create creates a new workout
func (s *WorkoutService) Create(ctx context.Context, workout Workout) (*Workout, error) {
	jsonBody, err := json.Marshal(workout)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal workout",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(ctx, "/workout-service/workout", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to create workout",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read workout response",
			Cause:      err,
		}
	}

	var createdWorkout Workout
	if err := json.Unmarshal(body, &createdWorkout); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse workout data",
			Cause:      err,
		}
	}

	return &createdWorkout, nil
}

// Update updates an existing workout
func (s *WorkoutService) Update(ctx context.Context, id string, update WorkoutUpdate) (*Workout, error) {
	jsonBody, err := json.Marshal(update)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal workout update",
			Cause:      err,
		}
	}

	path := "/workout-service/workout/" + id
	resp, err := s.client.Put(ctx, path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to update workout",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read workout response",
			Cause:      err,
		}
	}

	var updatedWorkout Workout
	if err := json.Unmarshal(body, &updatedWorkout); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse workout data",
			Cause:      err,
		}
	}

	return &updatedWorkout, nil
}

// Delete deletes an existing workout
func (s *WorkoutService) Delete(ctx context.Context, id string) error {
	path := "/workout-service/workout/" + id
	resp, err := s.client.Delete(ctx, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to delete workout",
		}
	}

	return nil
}

// GetWorkoutTemplates retrieves all workout templates for the current user
func (s *WorkoutService) GetWorkoutTemplates(ctx context.Context) ([]Workout, error) {
	path := "/workout-service/templates"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get workout templates",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read workout templates response",
			Cause:      err,
		}
	}

	var templates []Workout
	if err := json.Unmarshal(body, &templates); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse workout templates data",
			Cause:      err,
		}
	}

	return templates, nil
}

// SearchWorkouts searches workouts by name or description
func (s *WorkoutService) SearchWorkouts(ctx context.Context, query string, limit int) ([]Workout, error) {
	params := url.Values{}
	params.Set("q", query)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := "/workout-service/search?" + params.Encode()

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to search workouts",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read search response",
			Cause:      err,
		}
	}

	var workouts []Workout
	if err := json.Unmarshal(body, &workouts); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse search results",
			Cause:      err,
		}
	}

	return workouts, nil
}

// CopyWorkout creates a copy of an existing workout
func (s *WorkoutService) CopyWorkout(ctx context.Context, id string, newName string) (*Workout, error) {
	path := "/workout-service/workout/" + id + "/copy"

	requestBody := map[string]string{"name": newName}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to copy workout",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read copy response",
			Cause:      err,
		}
	}

	var copiedWorkout Workout
	if err := json.Unmarshal(body, &copiedWorkout); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse copied workout data",
			Cause:      err,
		}
	}

	return &copiedWorkout, nil
}

// Export exports a workout in the specified format (fit, tcx, json)
func (s *WorkoutService) Export(ctx context.Context, id string, format string) (io.ReadCloser, error) {
	if format != "fit" && format != "tcx" && format != "json" {
		return nil, &APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid format. Supported formats: fit, tcx, json",
		}
	}

	path := "/workout-service/workout/" + id + "/export/" + format

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to export workout",
		}
	}

	return resp.Body, nil
}
