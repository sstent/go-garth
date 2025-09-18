package stats

import (
	"errors"
	"time"
)

const WEEKLY_STRESS_PATH = "/wellness-service/wellness/weeklyStress"

type WeeklyStress struct {
	CalendarDate        time.Time `json:"calendar_date"`
	TotalStressDuration int       `json:"total_stress_duration"`
	AverageStressLevel  float64   `json:"average_stress_level"`
	MaxStressLevel      int       `json:"max_stress_level"`
	StressQualifier     string    `json:"stress_qualifier"`
	BaseStats
}

func NewWeeklyStress() *WeeklyStress {
	return &WeeklyStress{
		BaseStats: BaseStats{
			Path:     WEEKLY_STRESS_PATH + "/{end}/{period}",
			PageSize: 52,
		},
	}
}

func (w *WeeklyStress) Validate() error {
	if w.CalendarDate.IsZero() {
		return errors.New("calendar_date is required")
	}
	if w.TotalStressDuration < 0 {
		return errors.New("total_stress_duration must be non-negative")
	}
	return nil
}
