package data

import (
	types "go-garth/internal/models/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBodyBatteryReadings(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]any
		expected []BodyBatteryReading
	}{
		{
			name: "valid readings",
			input: [][]any{
				{1000, "ACTIVE", 75, 1.0},
				{2000, "ACTIVE", 70, 1.0},
				{3000, "REST", 65, 1.0},
			},
			expected: []BodyBatteryReading{
				{1000, "ACTIVE", 75, 1.0},
				{2000, "ACTIVE", 70, 1.0},
				{3000, "REST", 65, 1.0},
			},
		},
		{
			name: "invalid readings",
			input: [][]any{
				{1000, "ACTIVE", 75},           // missing version
				{2000, "ACTIVE"},               // missing level and version
				{3000},                         // only timestamp
				{"invalid", "ACTIVE", 75, 1.0}, // wrong timestamp type
			},
			expected: []BodyBatteryReading{},
		},
		{
			name:     "empty input",
			input:    [][]any{},
			expected: []BodyBatteryReading{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseBodyBatteryReadings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test for GetCurrentLevel and GetDayChange methods
func TestBodyBatteryDataWithMethods(t *testing.T) {
	mockData := types.DetailedBodyBatteryData{
		BodyBatteryValuesArray: [][]interface{}{
			{1000, "ACTIVE", 75, 1.0},
			{2000, "ACTIVE", 70, 1.0},
			{3000, "REST", 65, 1.0},
		},
	}

	bb := BodyBatteryDataWithMethods{DetailedBodyBatteryData: mockData}

	t.Run("GetCurrentLevel", func(t *testing.T) {
		assert.Equal(t, 65, bb.GetCurrentLevel())
	})

	t.Run("GetDayChange", func(t *testing.T) {
		assert.Equal(t, -10, bb.GetDayChange()) // 65 - 75 = -10
	})

	// Test with empty data
	emptyData := types.DetailedBodyBatteryData{
		BodyBatteryValuesArray: [][]interface{}{},
	}
	emptyBb := BodyBatteryDataWithMethods{DetailedBodyBatteryData: emptyData}

	t.Run("GetCurrentLevel empty", func(t *testing.T) {
		assert.Equal(t, 0, emptyBb.GetCurrentLevel())
	})

	t.Run("GetDayChange empty", func(t *testing.T) {
		assert.Equal(t, 0, emptyBb.GetDayChange())
	})

	// Test with single reading
	singleReadingData := types.DetailedBodyBatteryData{
		BodyBatteryValuesArray: [][]interface{}{
			{1000, "ACTIVE", 80, 1.0},
		},
	}
	singleReadingBb := BodyBatteryDataWithMethods{DetailedBodyBatteryData: singleReadingData}

	t.Run("GetDayChange single reading", func(t *testing.T) {
		assert.Equal(t, 0, singleReadingBb.GetDayChange())
	})
}
