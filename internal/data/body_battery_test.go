package data

import (
	"testing"
	"time"

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

func TestParseStressReadings(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]int
		expected []StressReading
	}{
		{
			name: "valid readings",
			input: [][]int{
				{1000, 25},
				{2000, 30},
				{3000, 20},
			},
			expected: []StressReading{
				{1000, 25},
				{2000, 30},
				{3000, 20},
			},
		},
		{
			name: "invalid readings",
			input: [][]int{
				{1000},        // missing stress level
				{2000, 30, 1}, // extra value
				{},            // empty
			},
			expected: []StressReading{},
		},
		{
			name:     "empty input",
			input:    [][]int{},
			expected: []StressReading{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStressReadings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDailyBodyBatteryStress(t *testing.T) {
	now := time.Now()
	d := DailyBodyBatteryStress{
		CalendarDate: now,
		BodyBatteryValuesArray: [][]any{
			{1000, "ACTIVE", 75, 1.0},
			{2000, "ACTIVE", 70, 1.0},
		},
		StressValuesArray: [][]int{
			{1000, 25},
			{2000, 30},
		},
	}

	t.Run("body battery readings", func(t *testing.T) {
		readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
		assert.Len(t, readings, 2)
		assert.Equal(t, 75, readings[0].Level)
	})

	t.Run("stress readings", func(t *testing.T) {
		readings := ParseStressReadings(d.StressValuesArray)
		assert.Len(t, readings, 2)
		assert.Equal(t, 25, readings[0].StressLevel)
	})
}
