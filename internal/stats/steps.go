package stats

import "time"

const BASE_STEPS_PATH = "/usersummary-service/stats/steps"

type DailySteps struct {
	CalendarDate  time.Time `json:"calendar_date"`
	TotalSteps    *int      `json:"total_steps"`
	TotalDistance *int      `json:"total_distance"`
	StepGoal      int       `json:"step_goal"`
	BaseStats
}

func NewDailySteps() *DailySteps {
	return &DailySteps{
		BaseStats: BaseStats{
			Path:     BASE_STEPS_PATH + "/daily/{start}/{end}",
			PageSize: 28,
		},
	}
}

type WeeklySteps struct {
	CalendarDate          time.Time `json:"calendar_date"`
	TotalSteps            int       `json:"total_steps"`
	AverageSteps          float64   `json:"average_steps"`
	AverageDistance       float64   `json:"average_distance"`
	TotalDistance         float64   `json:"total_distance"`
	WellnessDataDaysCount int       `json:"wellness_data_days_count"`
	BaseStats
}

func NewWeeklySteps() *WeeklySteps {
	return &WeeklySteps{
		BaseStats: BaseStats{
			Path:     BASE_STEPS_PATH + "/weekly/{end}/{period}",
			PageSize: 52,
		},
	}
}
