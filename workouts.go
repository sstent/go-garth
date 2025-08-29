package garth

import (
	"bytes"
	"context"
	encjson "encoding/json"
	"io"
	"net/http"
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

// CreateWorkout creates a new workout in Garmin Connect.
// workout: The workout data to create.
// Returns:
//   - *Workout: The created workout data
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) CreateWorkout(workout *Workout) (*Workout, *http.Response, error) {
	path := "/workout-service/workout"

	jsonBody, err := encjson.Marshal(workout)
	if err != nil {
		return nil, nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(context.Background(), path, bytes.NewReader(jsonBody))
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, nil, apiErr
		}
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var createdWorkout Workout
	if err := encjson.NewDecoder(resp.Body).Decode(&createdWorkout); err != nil {
		return nil, resp, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse response",
			Cause:      err,
		}
	}

	return &createdWorkout, resp, nil
}

// GetWorkout retrieves a specific workout by ID.
// id: The ID of the workout to retrieve.
// Returns:
//   - *Workout: The retrieved workout data
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) GetWorkout(id string) (*Workout, *http.Response, error) {
	path := "/workout-service/workout/" + id

	var workout Workout
	resp, err := s.client.Get(context.Background(), path)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, nil, apiErr
		}
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if err := encjson.NewDecoder(resp.Body).Decode(&workout); err != nil {
		return nil, resp, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse response",
			Cause:      err,
		}
	}

	return &workout, resp, nil
}

// UpdateWorkout modifies an existing workout.
// id: The ID of the workout to update.
// workout: The updated workout data.
// Returns:
//   - *Workout: The updated workout data
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) UpdateWorkout(id string, workout *Workout) (*Workout, *http.Response, error) {
	path := "/workout-service/workout/" + id

	jsonBody, err := encjson.Marshal(workout)
	if err != nil {
		return nil, nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := s.client.Put(context.Background(), path, bytes.NewReader(jsonBody))
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, nil, apiErr
		}
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var updatedWorkout Workout
	if err := encjson.NewDecoder(resp.Body).Decode(&updatedWorkout); err != nil {
		return nil, resp, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse response",
			Cause:      err,
		}
	}

	return &updatedWorkout, resp, nil
}

// DeleteWorkout removes a workout.
// id: The ID of the workout to delete.
// Returns:
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) DeleteWorkout(id string) (*http.Response, error) {
	path := "/workout-service/workout/" + id

	resp, err := s.client.Delete(context.Background(), path, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, apiErr
		}
		return resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return resp, nil
}

// ScheduleWorkout schedules a workout for a specific date.
// id: The ID of the workout to schedule.
// date: The date to schedule the workout (format: YYYY-MM-DD).
// Returns:
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) ScheduleWorkout(id string, date string) (*http.Response, error) {
	path := "/workout-service/schedule/" + id

	// Create request body
	requestBody := map[string]string{"date": date}
	jsonBody, err := encjson.Marshal(requestBody)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(context.Background(), path, bytes.NewReader(jsonBody))
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, apiErr
		}
		return resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return resp, nil
}

// CreateWorkoutTemplate creates a reusable workout template.
// workout: The workout data to create as a template.
// Returns:
//   - *Workout: The created template data
//   - *http.Response: The HTTP response from Garmin
//   - error: Any error that occurred
func (s *WorkoutService) CreateWorkoutTemplate(workout *Workout) (*Workout, *http.Response, error) {
	path := "/workout-service/template"

	jsonBody, err := encjson.Marshal(workout)
	if err != nil {
		return nil, nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(context.Background(), path, bytes.NewReader(jsonBody))
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return nil, nil, apiErr
		}
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var createdTemplate Workout
	if err := encjson.NewDecoder(resp.Body).Decode(&createdTemplate); err != nil {
		return nil, resp, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse response",
			Cause:      err,
		}
	}

	return &createdTemplate, resp, nil
}

// Workout represents Garmin workout data.
// Additional fields will be added as needed to match the Garmin API.
type Workout struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Additional fields will be added during implementation
}
