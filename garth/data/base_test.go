package data

import (
	"errors"
	"testing"
	"time"

	"garmin-connect/garth/client"

	"github.com/stretchr/testify/assert"
)

// MockData implements Data interface for testing
type MockData struct {
	BaseData
}

// MockClient simulates API client for tests
type MockClient struct{}

func (mc *MockClient) Get(endpoint string) (interface{}, error) {
	if endpoint == "error" {
		return nil, errors.New("mock API error")
	}
	return "data for " + endpoint, nil
}

func TestBaseData_List(t *testing.T) {
	// Setup mock data type
	mockData := &MockData{}
	mockData.GetFunc = func(day time.Time, c *client.Client) (interface{}, error) {
		return "data for " + day.Format("2006-01-02"), nil
	}

	// Test parameters
	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	days := 5
	c := &client.Client{}
	maxWorkers := 3

	// Execute
	results, errs := mockData.List(end, days, c, maxWorkers)

	// Verify
	assert.Empty(t, errs)
	assert.Len(t, results, days)
	assert.Contains(t, results, "data for 2023-06-15")
	assert.Contains(t, results, "data for 2023-06-11")
}

func TestBaseData_List_ErrorHandling(t *testing.T) {
	// Setup mock data type that returns error on specific date
	mockData := &MockData{}
	mockData.GetFunc = func(day time.Time, c *client.Client) (interface{}, error) {
		if day.Day() == 13 {
			return nil, errors.New("bad luck day")
		}
		return "data for " + day.Format("2006-01-02"), nil
	}

	// Test parameters
	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	days := 5
	c := &client.Client{}
	maxWorkers := 2

	// Execute
	results, errs := mockData.List(end, days, c, maxWorkers)

	// Verify
	assert.Len(t, errs, 1)
	assert.Equal(t, "bad luck day", errs[0].Error())
	assert.Len(t, results, 4) // Should have results for non-error days
}
