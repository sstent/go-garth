package stats

import "time"

const BASE_SLEEP_PATH = "/usersummary-service/stats/sleep"

type DailySleep struct {
	CalendarDate        time.Time `json:"calendar_date"`
	TotalSleepTime      *int      `json:"total_sleep_time"`
	RemSleepTime        *int      `json:"rem_sleep_time"`
	DeepSleepTime       *int      `json:"deep_sleep_time"`
	LightSleepTime      *int      `json:"light_sleep_time"`
	AwakeTime           *int      `json:"awake_time"`
	SleepScore          *int      `json:"sleep_score"`
	SleepStartTimestamp *int64    `json:"sleep_start_timestamp"`
	SleepEndTimestamp   *int64    `json:"sleep_end_timestamp"`
	BaseStats
}

func NewDailySleep() *DailySleep {
	return &DailySleep{
		BaseStats: BaseStats{
			Path:     BASE_SLEEP_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}
