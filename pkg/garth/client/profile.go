package client

import (
	"time"
)

type UserProfile struct {
	ID                            int       `json:"id"`
	ProfileID                     int       `json:"profileId"`
	GarminGUID                    string    `json:"garminGuid"`
	DisplayName                   string    `json:"displayName"`
	FullName                      string    `json:"fullName"`
	UserName                      string    `json:"userName"`
	ProfileImageType              *string   `json:"profileImageType"`
	ProfileImageURLLarge          *string   `json:"profileImageUrlLarge"`
	ProfileImageURLMedium         *string   `json:"profileImageUrlMedium"`
	ProfileImageURLSmall          *string   `json:"profileImageUrlSmall"`
	Location                      *string   `json:"location"`
	FacebookURL                   *string   `json:"facebookUrl"`
	TwitterURL                    *string   `json:"twitterUrl"`
	PersonalWebsite               *string   `json:"personalWebsite"`
	Motivation                    *string   `json:"motivation"`
	Bio                           *string   `json:"bio"`
	PrimaryActivity               *string   `json:"primaryActivity"`
	FavoriteActivityTypes         []string  `json:"favoriteActivityTypes"`
	RunningTrainingSpeed          float64   `json:"runningTrainingSpeed"`
	CyclingTrainingSpeed          float64   `json:"cyclingTrainingSpeed"`
	FavoriteCyclingActivityTypes  []string  `json:"favoriteCyclingActivityTypes"`
	CyclingClassification         *string   `json:"cyclingClassification"`
	CyclingMaxAvgPower            float64   `json:"cyclingMaxAvgPower"`
	SwimmingTrainingSpeed         float64   `json:"swimmingTrainingSpeed"`
	ProfileVisibility             string    `json:"profileVisibility"`
	ActivityStartVisibility       string    `json:"activityStartVisibility"`
	ActivityMapVisibility         string    `json:"activityMapVisibility"`
	CourseVisibility              string    `json:"courseVisibility"`
	ActivityHeartRateVisibility   string    `json:"activityHeartRateVisibility"`
	ActivityPowerVisibility       string    `json:"activityPowerVisibility"`
	BadgeVisibility               string    `json:"badgeVisibility"`
	ShowAge                       bool      `json:"showAge"`
	ShowWeight                    bool      `json:"showWeight"`
	ShowHeight                    bool      `json:"showHeight"`
	ShowWeightClass               bool      `json:"showWeightClass"`
	ShowAgeRange                  bool      `json:"showAgeRange"`
	ShowGender                    bool      `json:"showGender"`
	ShowActivityClass             bool      `json:"showActivityClass"`
	ShowVO2Max                    bool      `json:"showVo2Max"`
	ShowPersonalRecords           bool      `json:"showPersonalRecords"`
	ShowLast12Months              bool      `json:"showLast12Months"`
	ShowLifetimeTotals            bool      `json:"showLifetimeTotals"`
	ShowUpcomingEvents            bool      `json:"showUpcomingEvents"`
	ShowRecentFavorites           bool      `json:"showRecentFavorites"`
	ShowRecentDevice              bool      `json:"showRecentDevice"`
	ShowRecentGear                bool      `json:"showRecentGear"`
	ShowBadges                    bool      `json:"showBadges"`
	OtherActivity                 *string   `json:"otherActivity"`
	OtherPrimaryActivity          *string   `json:"otherPrimaryActivity"`
	OtherMotivation               *string   `json:"otherMotivation"`
	UserRoles                     []string  `json:"userRoles"`
	NameApproved                  bool      `json:"nameApproved"`
	UserProfileFullName           string    `json:"userProfileFullName"`
	MakeGolfScorecardsPrivate     bool      `json:"makeGolfScorecardsPrivate"`
	AllowGolfLiveScoring          bool      `json:"allowGolfLiveScoring"`
	AllowGolfScoringByConnections bool      `json:"allowGolfScoringByConnections"`
	UserLevel                     int       `json:"userLevel"`
	UserPoint                     int       `json:"userPoint"`
	LevelUpdateDate               time.Time `json:"levelUpdateDate"`
	LevelIsViewed                 bool      `json:"levelIsViewed"`
	LevelPointThreshold           int       `json:"levelPointThreshold"`
	UserPointOffset               int       `json:"userPointOffset"`
	UserPro                       bool      `json:"userPro"`
}
