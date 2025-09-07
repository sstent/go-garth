package data

import (
	"errors"
	"sync"
	"time"

	"garmin-connect/garth/client"
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
		maxWorkers = 1
	}

	// Generate date range (end backwards for 'days' days)
	dates := make([]time.Time, days)
	for i := 0; i < days; i++ {
		dates[i] = end.AddDate(0, 0, -i)
	}

	var wg sync.WaitGroup
	workCh := make(chan time.Time, days)
	resultsCh := make(chan interface{}, days)
	errCh := make(chan error, days)

	// Worker function
	worker := func() {
		defer wg.Done()
		for date := range workCh {
			result, err := b.Get(date, c)
			if err != nil {
				errCh <- err
				continue
			}
			resultsCh <- result
		}
	}

	// Start workers
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Send work to channel
	go func() {
		for _, date := range dates {
			workCh <- date
		}
		close(workCh)
	}()

	// Close channels when all workers finish
	go func() {
		wg.Wait()
		close(resultsCh)
		close(errCh)
	}()

	// Collect results and errors
	var results []interface{}
	var errs []error

	// Collect results until both channels are closed
	for resultsCh != nil || errCh != nil {
		select {
		case result, ok := <-resultsCh:
			if !ok {
				resultsCh = nil
				continue
			}
			results = append(results, result)
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			errs = append(errs, err)
		}
	}

	return results, errs
}
