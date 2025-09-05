package garth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestWorkoutService_List tests the List method with various options
func TestWorkoutService_List(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []Workout
		mockStatusCode int
		opts           WorkoutListOptions
		wantErr        bool
	}{
		{
			name: "successful list with no options",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
				{WorkoutID: 2, Name: "Evening Ride", Type: "cycling"},
			},
			mockStatusCode: http.StatusOK,
			opts:           WorkoutListOptions{},
			wantErr:        false,
		},
		{
			name: "successful list with limit",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
			},
			mockStatusCode: http.StatusOK,
			opts:           WorkoutListOptions{Limit: 1},
			wantErr:        false,
		},
		{
			name: "successful list with date range",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
			},
			mockStatusCode: http.StatusOK,
			opts: WorkoutListOptions{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "successful list with pagination",
			mockResponse: []Workout{
				{WorkoutID: 2, Name: "Evening Ride", Type: "cycling"},
			},
			mockStatusCode: http.StatusOK,
			opts: WorkoutListOptions{
				Limit:  1,
				Offset: 1,
			},
			wantErr: false,
		},
		{
			name: "successful list with sorting",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
				{WorkoutID: 2, Name: "Evening Ride", Type: "cycling"},
			},
			mockStatusCode: http.StatusOK,
			opts: WorkoutListOptions{
				SortBy:    "createdDate",
				SortOrder: "desc",
			},
			wantErr: false,
		},
		{
			name: "successful list with type filter",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
			},
			mockStatusCode: http.StatusOK,
			opts: WorkoutListOptions{
				Type: "running",
			},
			wantErr: false,
		},
		{
			name: "successful list with status filter",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
			},
			mockStatusCode: http.StatusOK,
			opts: WorkoutListOptions{
				Status: "active",
			},
			wantErr: false,
		},
		{
			name:           "server error",
			mockResponse:   nil,
			mockStatusCode: http.StatusInternalServerError,
			opts:           WorkoutListOptions{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/workout-service/workouts" {
					t.Errorf("expected path /workout-service/workouts, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			workouts, err := service.List(context.Background(), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(workouts) != len(tt.mockResponse) {
				t.Errorf("WorkoutService.List() got %d workouts, want %d", len(workouts), len(tt.mockResponse))
			}
		})
	}
}

// TestWorkoutService_Get tests the Get method
func TestWorkoutService_Get(t *testing.T) {
	tests := []struct {
		name           string
		workoutID      string
		mockResponse   *WorkoutDetails
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful get",
			workoutID: "123",
			mockResponse: &WorkoutDetails{
				Workout: Workout{
					WorkoutID: 123,
					Name:      "Test Workout",
					Type:      "running",
				},
				EstimatedDuration: 3600,
				TrainingLoad:      50.5,
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "workout not found",
			workoutID:      "999",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/workout-service/workout/" + tt.workoutID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			workout, err := service.Get(context.Background(), tt.workoutID)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && workout.WorkoutID != 123 {
				t.Errorf("WorkoutService.Get() got ID %d, want 123", workout.WorkoutID)
			}
		})
	}
}

// TestWorkoutService_Create tests the Create method
func TestWorkoutService_Create(t *testing.T) {
	tests := []struct {
		name           string
		workout        Workout
		mockResponse   *Workout
		mockStatusCode int
		wantErr        bool
	}{
		{
			name: "successful create",
			workout: Workout{
				Name:        "New Workout",
				Description: "Test workout",
				Type:        "cycling",
			},
			mockResponse: &Workout{
				WorkoutID:   456,
				Name:        "New Workout",
				Description: "Test workout",
				Type:        "cycling",
			},
			mockStatusCode: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "invalid workout data",
			workout: Workout{
				Name: "",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/workout-service/workout" {
					t.Errorf("expected path /workout-service/workout, got %s", r.URL.Path)
				}

				body, _ := ioutil.ReadAll(r.Body)
				var receivedWorkout Workout
				json.Unmarshal(body, &receivedWorkout)

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			createdWorkout, err := service.Create(context.Background(), tt.workout)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && createdWorkout.Name != tt.workout.Name {
				t.Errorf("WorkoutService.Create() got name %s, want %s", createdWorkout.Name, tt.workout.Name)
			}
		})
	}
}

// TestWorkoutService_Update tests the Update method
func TestWorkoutService_Update(t *testing.T) {
	tests := []struct {
		name           string
		workoutID      string
		update         WorkoutUpdate
		mockResponse   *Workout
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful update",
			workoutID: "123",
			update: WorkoutUpdate{
				Name:        "Updated Workout",
				Description: "Updated description",
			},
			mockResponse: &Workout{
				WorkoutID:   123,
				Name:        "Updated Workout",
				Description: "Updated description",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:      "workout not found",
			workoutID: "999",
			update: WorkoutUpdate{
				Name: "Updated Workout",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/workout-service/workout/" + tt.workoutID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				body, _ := ioutil.ReadAll(r.Body)
				var receivedUpdate WorkoutUpdate
				json.Unmarshal(body, &receivedUpdate)

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			updatedWorkout, err := service.Update(context.Background(), tt.workoutID, tt.update)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && updatedWorkout.Name != tt.update.Name {
				t.Errorf("WorkoutService.Update() got name %s, want %s", updatedWorkout.Name, tt.update.Name)
			}
		})
	}
}

// TestWorkoutService_Delete tests the Delete method
func TestWorkoutService_Delete(t *testing.T) {
	tests := []struct {
		name           string
		workoutID      string
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "successful delete",
			workoutID:      "123",
			mockStatusCode: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "workout not found",
			workoutID:      "999",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/workout-service/workout/" + tt.workoutID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			err := service.Delete(context.Background(), tt.workoutID)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWorkoutService_SearchWorkouts tests the SearchWorkouts method
func TestWorkoutService_SearchWorkouts(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          int
		mockResponse   []Workout
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:  "successful search",
			query: "running",
			limit: 10,
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Morning Run", Type: "running"},
				{WorkoutID: 2, Name: "Evening Run", Type: "running"},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty search results",
			query:          "nonexistent",
			limit:          10,
			mockResponse:   []Workout{},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/workout-service/search" {
					t.Errorf("expected path /workout-service/search, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			workouts, err := service.SearchWorkouts(context.Background(), tt.query, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.SearchWorkouts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(workouts) != len(tt.mockResponse) {
				t.Errorf("WorkoutService.SearchWorkouts() got %d workouts, want %d", len(workouts), len(tt.mockResponse))
			}
		})
	}
}

// TestWorkoutService_GetWorkoutTemplates tests the GetWorkoutTemplates method
func TestWorkoutService_GetWorkoutTemplates(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []Workout
		mockStatusCode int
		wantErr        bool
	}{
		{
			name: "successful get templates",
			mockResponse: []Workout{
				{WorkoutID: 1, Name: "Template 1", Type: "running"},
				{WorkoutID: 2, Name: "Template 2", Type: "cycling"},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "no templates",
			mockResponse:   []Workout{},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/workout-service/templates" {
					t.Errorf("expected path /workout-service/templates, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			templates, err := service.GetWorkoutTemplates(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.GetWorkoutTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(templates) != len(tt.mockResponse) {
				t.Errorf("WorkoutService.GetWorkoutTemplates() got %d templates, want %d", len(templates), len(tt.mockResponse))
			}
		})
	}
}

// TestWorkoutService_CopyWorkout tests the CopyWorkout method
func TestWorkoutService_CopyWorkout(t *testing.T) {
	tests := []struct {
		name           string
		workoutID      string
		newName        string
		mockResponse   *Workout
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful copy",
			workoutID: "123",
			newName:   "Copied Workout",
			mockResponse: &Workout{
				WorkoutID: 456,
				Name:      "Copied Workout",
			},
			mockStatusCode: http.StatusCreated,
			wantErr:        false,
		},
		{
			name:           "workout not found",
			workoutID:      "999",
			newName:        "Copied Workout",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/workout-service/workout/" + tt.workoutID + "/copy"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				body, _ := ioutil.ReadAll(r.Body)
				var requestBody map[string]string
				json.Unmarshal(body, &requestBody)

				if requestBody["name"] != tt.newName {
					t.Errorf("expected name %s, got %s", tt.newName, requestBody["name"])
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			copiedWorkout, err := service.CopyWorkout(context.Background(), tt.workoutID, tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.CopyWorkout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && copiedWorkout.Name != tt.newName {
				t.Errorf("WorkoutService.CopyWorkout() got name %s, want %s", copiedWorkout.Name, tt.newName)
			}
		})
	}
}

// TestWorkoutService_Export tests the Export method
func TestWorkoutService_Export(t *testing.T) {
	tests := []struct {
		name           string
		workoutID      string
		format         string
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "successful export fit",
			workoutID:      "123",
			format:         "fit",
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "successful export tcx",
			workoutID:      "123",
			format:         "tcx",
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "invalid format",
			workoutID:      "123",
			format:         "invalid",
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "workout not found",
			workoutID:      "999",
			format:         "fit",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/download-service/export/" + tt.format + "/workout/" + tt.workoutID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if !tt.wantErr {
					w.Write([]byte("test export data"))
				}
			}))
			defer server.Close()

			client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
			service := NewWorkoutService(client)

			reader, err := service.Export(context.Background(), tt.workoutID, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkoutService.Export() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer reader.Close()
				data, _ := ioutil.ReadAll(reader)
				if string(data) != "test export data" {
					t.Errorf("WorkoutService.Export() got unexpected export data")
				}
			}
		})
	}
}

// TestWorkoutService_ContextMethods tests the new context-aware methods
func TestWorkoutService_ContextMethods(t *testing.T) {
	t.Run("Create with context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(&Workout{WorkoutID: 123, Name: "Test Workout"})
		}))
		defer server.Close()

		client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
		service := NewWorkoutService(client)

		workout, err := service.Create(context.Background(), Workout{Name: "Test Workout"})
		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}

		if workout.WorkoutID != 123 {
			t.Errorf("Create() got ID %d, want 123", workout.WorkoutID)
		}
	})

	t.Run("Get with context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&WorkoutDetails{
				Workout: Workout{WorkoutID: 123, Name: "Test Workout"},
			})
		}))
		defer server.Close()

		client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
		service := NewWorkoutService(client)

		workout, err := service.Get(context.Background(), "123")
		if err != nil {
			t.Errorf("Get() error = %v", err)
			return
		}

		if workout.WorkoutID != 123 {
			t.Errorf("Get() got ID %d, want 123", workout.WorkoutID)
		}
	})

	t.Run("Update with context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&Workout{WorkoutID: 123, Name: "Updated Workout"})
		}))
		defer server.Close()

		client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
		service := NewWorkoutService(client)

		updatedWorkout, err := service.Update(context.Background(), "123", WorkoutUpdate{Name: "Updated Workout"})
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}

		if updatedWorkout.Name != "Updated Workout" {
			t.Errorf("Update() got name %s, want 'Updated Workout'", updatedWorkout.Name)
		}
	})

	t.Run("Delete with context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &APIClient{baseURL: server.URL, httpClient: server.Client()}
		service := NewWorkoutService(client)

		err := service.Delete(context.Background(), "123")
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})
}
