package stats

import "time"

const BASE_INTENSITY_PATH = "/usersummary-service/stats/intensity_minutes"

type DailyIntensityMinutes struct {
	CalendarDate      time.Time `json:"calendar_date"`
	ModerateIntensity *int      `json:"moderate_intensity"`
	VigorousIntensity *int      `json:"vigorous_intensity"`
	BaseStats
}

func NewDailyIntensityMinutes() *DailyIntensityMinutes {
	return &DailyIntensityMinutes{
		BaseStats: BaseStats{
			Path:     BASE_INTENSITY_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}
