package data

import (
	"errors"
	"sync"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/utils"
)

// Data defines the interface for Garmin Connect data types.
// Concrete data types (BodyBattery, HRV, Sleep, etc.) must implement this interface.
//
// The Get method retrieves data for a single day.
// The List method concurrently retrieves data for a range of days.
type Data interface {
	Get(day time.Time, c *client.Client) (interface{}, error)
	List(end time.Time, days int, c *client.Client, maxWorkers int) ([]interface{}, error)
}

// BaseData provides a reusable implementation for data types to embed.
// It handles the concurrent List() implementation while allowing concrete types
// to focus on implementing the Get() method for their specific data structure.
//
// Usage:
//
//	type BodyBatteryData struct {
//	    data.BaseData
//	    // ... additional fields
//	}
//
//	func NewBodyBatteryData() *BodyBatteryData {
//	    bb := &BodyBatteryData{}
//	    bb.GetFunc = bb.get // Assign the concrete Get implementation
//	    return bb
//	}
//
//	func (bb *BodyBatteryData) get(day time.Time, c *client.Client) (interface{}, error) {
//	    // Implementation specific to body battery data
//	}
type BaseData struct {
	// GetFunc must be set by concrete types to implement the Get method.
	// This function pointer allows BaseData to call the concrete implementation.
	GetFunc func(day time.Time, c *client.Client) (interface{}, error)
}

// Get implements the Data interface by calling the configured GetFunc.
// Returns an error if GetFunc is not set.
func (b *BaseData) Get(day time.Time, c *client.Client) (interface{}, error) {
	if b.GetFunc == nil {
		return nil, errors.New("GetFunc not implemented for this data type")
	}
	return b.GetFunc(day, c)
}

// List implements concurrent data fetching using a worker pool pattern.
// This method efficiently retrieves data for multiple days by distributing
// work across a configurable number of workers (goroutines).
//
// Parameters:
//
//	end: The end date of the range (inclusive)
//	days: Number of days to fetch (going backwards from end date)
//	c: Client instance for API access
//	maxWorkers: Maximum concurrent workers (minimum 1)
//
// Returns:
//
//	[]interface{}: Slice of results (order matches date range)
//	[]error: Slice of errors encountered during processing
func (b *BaseData) List(end time.Time, days int, c *client.Client, maxWorkers int) ([]interface{}, []error) {
	if maxWorkers < 1 {
		maxWorkers = 10 // Match Python's MAX_WORKERS
	}

	dates := utils.DateRange(end, days)

	// Define result type for channel
	type result struct {
		data interface{}
		err  error
	}

	var wg sync.WaitGroup
	workCh := make(chan time.Time, days)
	resultsCh := make(chan result, days)

	// Worker function
	worker := func() {
		defer wg.Done()
		for date := range workCh {
			data, err := b.Get(date, c)
			resultsCh <- result{data: data, err: err}
		}
	}

	// Start workers
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Send work
	go func() {
		for _, date := range dates {
			workCh <- date
		}
		close(workCh)
	}()

	// Close results channel when workers are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []interface{}
	var errs []error

	for r := range resultsCh {
		if r.err != nil {
			errs = append(errs, r.err)
		} else if r.data != nil {
			results = append(results, r.data)
		}
	}

	return results, errs
}
