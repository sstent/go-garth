package stats

import (
	"fmt"
	"strings"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/utils"
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

	if period > b.PageSize {
		// Handle pagination - get first page
		page, err := b.fetchPage(endDate, b.PageSize, client)
		if err != nil || len(page) == 0 {
			return page, err
		}

		// Get remaining pages recursively
		remainingStart := endDate.AddDate(0, 0, -b.PageSize)
		remainingPeriod := period - b.PageSize
		remainingData, err := b.List(remainingStart, remainingPeriod, client)
		if err != nil {
			return page, err
		}

		return append(remainingData, page...), nil
	}

	return b.fetchPage(endDate, period, client)
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

	response, err := client.ConnectAPI(path, "GET", nil)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return []interface{}{}, nil
	}

	responseSlice, ok := response.([]interface{})
	if !ok || len(responseSlice) == 0 {
		return []interface{}{}, nil
	}

	var results []interface{}
	for _, item := range responseSlice {
		itemMap := item.(map[string]interface{})

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
