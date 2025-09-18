package activities

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/sstent/go-garth/garth/types"
	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

func (m *mockClient) GetActivities(start, end time.Time) ([]types.Activity, error) {
	return []types.Activity{
		{
			ActivityID:   123,
			ActivityName: "Morning Run",
			ActivityType: types.ActivityType{
				TypeKey: "running",
			},
			StartTimeLocal: "2025-09-18T08:30:00",
			Distance:       5000,
			Duration:       1800,
		},
	}, nil
}

func (m *mockClient) Login(email, password string) error {
	return nil
}

func (m *mockClient) SaveSession(filename string) error {
	return nil
}

func TestRenderTable(t *testing.T) {
	activities := []types.Activity{
		{
			ActivityID:   123,
			ActivityName: "Morning Run",
			ActivityType: types.ActivityType{
				TypeKey: "running",
			},
			StartTimeLocal: "2025-09-18T08:30:00",
			Distance:       5000,
			Duration:       1800,
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderTable(activities)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "123")
	assert.Contains(t, output, "Morning Run")
	assert.Contains(t, output, "running")
}

func TestRenderJSON(t *testing.T) {
	activities := []types.Activity{
		{
			ActivityID:   123,
			ActivityName: "Morning Run",
			ActivityType: types.ActivityType{
				TypeKey: "running",
			},
			StartTimeLocal: "2025-09-18T08:30:00",
			Distance:       5000,
			Duration:       1800,
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := renderJSON(activities)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, `"activityId": 123`)
	assert.Contains(t, output, `"activityName": "Morning Run"`)
}

func TestRenderCSV(t *testing.T) {
	activities := []types.Activity{
		{
			ActivityID:   123,
			ActivityName: "Morning Run",
			ActivityType: types.ActivityType{
				TypeKey: "running",
			},
			StartTimeLocal: "2025-09-18T08:30:00",
			Distance:       5000,
			Duration:       1800,
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := renderCSV(activities)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "123,Morning Run,running")
}

func TestGetDateFilter(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantStart bool
		wantEnd   bool
	}{
		{
			name:      "no dates",
			startDate: "",
			endDate:   "",
			wantStart: false,
			wantEnd:   false,
		},
		{
			name:      "start date only",
			startDate: "2025-09-18",
			endDate:   "",
			wantStart: true,
			wantEnd:   false,
		},
		{
			name:      "both dates",
			startDate: "2025-09-18",
			endDate:   "2025-09-20",
			wantStart: true,
			wantEnd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origStart, origEnd := startDate, endDate
			defer func() {
				startDate, endDate = origStart, origEnd
			}()

			startDate = tt.startDate
			endDate = tt.endDate

			start, end := getDateFilter()

			if tt.wantStart {
				assert.False(t, start.IsZero(), "expected non-zero start time")
			} else {
				assert.True(t, start.IsZero(), "expected zero start time")
			}

			if tt.wantEnd {
				assert.False(t, end.IsZero(), "expected non-zero end time")
			} else {
				assert.True(t, end.IsZero(), "expected zero end time")
			}
		})
	}
}
