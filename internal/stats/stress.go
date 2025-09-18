package stats

import "time"

const BASE_STRESS_PATH = "/usersummary-service/stats/stress"

type DailyStress struct {
	CalendarDate         time.Time `json:"calendar_date"`
	OverallStressLevel   int       `json:"overall_stress_level"`
	RestStressDuration   *int      `json:"rest_stress_duration"`
	LowStressDuration    *int      `json:"low_stress_duration"`
	MediumStressDuration *int      `json:"medium_stress_duration"`
	HighStressDuration   *int      `json:"high_stress_duration"`
	BaseStats
}

func NewDailyStress() *DailyStress {
	return &DailyStress{
		BaseStats: BaseStats{
			Path:     BASE_STRESS_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}
