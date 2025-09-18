package stats

import "time"

const BASE_HYDRATION_PATH = "/usersummary-service/stats/hydration"

type DailyHydration struct {
	CalendarDate time.Time `json:"calendar_date"`
	TotalWaterML *int      `json:"total_water_ml"`
	BaseStats
}

func NewDailyHydration() *DailyHydration {
	return &DailyHydration{
		BaseStats: BaseStats{
			Path:     BASE_HYDRATION_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}
