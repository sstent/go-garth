package stats

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-garth/internal/api/client"
	"go-garth/internal/utils"
)

type Stats interface {
	List(end time.Time, period int, client *client.Client) ([]interface{}, error)
}

type BaseStats struct {
	Path     string
	PageSize int
}

func (b *BaseStats) List(end time.Time, period int, client *client.Client) ([]interface{}, error) {
	endDate := utils.FormatEndDate(end)
	var allData []interface{}
	var errs []error

	for period > 0 {
		pageSize := b.PageSize
		if period < pageSize {
			pageSize = period
		}

		page, err := b.fetchPage(endDate, pageSize, client)
		if err != nil {
			errs = append(errs, err)
			// Continue to next page even if current fails
		} else {
			allData = append(page, allData...)
		}

		// Move to previous page
		endDate = endDate.AddDate(0, 0, -pageSize)
		period -= pageSize
	}

	// Return partial data with aggregated errors
	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("partial failure: %v", errs)
	}
	return allData, finalErr
}

func (b *BaseStats) fetchPage(end time.Time, period int, client *client.Client) ([]interface{}, error) {
	var start time.Time
	var path string

	if strings.Contains(b.Path, "daily") {
		start = end.AddDate(0, 0, -(period - 1))
		path = strings.Replace(b.Path, "{start}", start.Format("2006-01-02"), 1)
		path = strings.Replace(path, "{end}", end.Format("2006-01-02"), 1)
	} else {
		path = strings.Replace(b.Path, "{end}", end.Format("2006-01-02"), 1)
		path = strings.Replace(path, "{period}", fmt.Sprintf("%d", period), 1)
	}

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []interface{}{}, nil
	}

	var responseSlice []map[string]interface{}
	if err := json.Unmarshal(data, &responseSlice); err != nil {
		return nil, err
	}

	if len(responseSlice) == 0 {
		return []interface{}{}, nil
	}

	var results []interface{}
	for _, itemMap := range responseSlice {
		// Handle nested "values" structure
		if values, exists := itemMap["values"]; exists {
			valuesMap := values.(map[string]interface{})
			for k, v := range valuesMap {
				itemMap[k] = v
			}
			delete(itemMap, "values")
		}

		snakeItem := utils.CamelToSnakeDict(itemMap)
		results = append(results, snakeItem)
	}

	return results, nil
}
