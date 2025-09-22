package garth

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGarminTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "space separated format",
			input:    `"2025-09-21 07:18:03"`,
			expected: time.Date(2025, 9, 21, 7, 18, 3, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "T separator with milliseconds",
			input:    `"2018-09-01T00:13:25.0"`,
			expected: time.Date(2018, 9, 1, 0, 13, 25, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "T separator without milliseconds",
			input:    `"2018-09-01T00:13:25"`,
			expected: time.Date(2018, 9, 1, 0, 13, 25, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "date only",
			input:    `"2018-09-01"`,
			expected: time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   `"invalid"`,
			wantErr: true,
		},
		{
			name:    "null value",
			input:   "null",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gt GarminTime
			err := json.Unmarshal([]byte(tt.input), &gt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.input == "null" {
				// For null values, the time should be zero
				if !gt.Time.IsZero() {
					t.Errorf("expected zero time for null input, got %v", gt.Time)
				}
				return
			}

			if !gt.Time.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, gt.Time)
			}
		})
	}
}
