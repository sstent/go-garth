package stats

import (
	"errors"
	"time"
)

const WEEKLY_HRV_PATH = "/wellness-service/wellness/weeklyHrv"

type WeeklyHRV struct {
	CalendarDate          time.Time `json:"calendar_date"`
	AverageHRV            float64   `json:"average_hrv"`
	MaxHRV                float64   `json:"max_hrv"`
	MinHRV                float64   `json:"min_hrv"`
	HRVQualifier          string    `json:"hrv_qualifier"`
	WellnessDataDaysCount int       `json:"wellness_data_days_count"`
	BaseStats
}

func NewWeeklyHRV() *WeeklyHRV {
	return &WeeklyHRV{
		BaseStats: BaseStats{
			Path:     WEEKLY_HRV_PATH + "/{end}/{period}",
			PageSize: 52,
		},
	}
}

func (w *WeeklyHRV) Validate() error {
	if w.CalendarDate.IsZero() {
		return errors.New("calendar_date is required")
	}
	if w.AverageHRV < 0 || w.MaxHRV < 0 || w.MinHRV < 0 {
		return errors.New("HRV values must be non-negative")
	}
	if w.MaxHRV < w.MinHRV {
		return errors.New("max_hrv must be greater than min_hrv")
	}
	return nil
}
