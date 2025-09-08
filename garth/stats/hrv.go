package stats

import "time"

const BASE_HRV_PATH = "/usersummary-service/stats/hrv"

type DailyHRV struct {
	CalendarDate time.Time `json:"calendar_date"`
	RestingHR    *int      `json:"resting_hr"`
	HRV          *int      `json:"hrv"`
	BaseStats
}

func NewDailyHRV() *DailyHRV {
	return &DailyHRV{
		BaseStats: BaseStats{
			Path:     BASE_HRV_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}
