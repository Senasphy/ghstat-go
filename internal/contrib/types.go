package contrib

import "time"

type Level string

const (
	LevelNone           Level = "NONE"
	LevelFirstQuartile  Level = "FIRST_QUARTILE"
	LevelSecondQuartile Level = "SECOND_QUARTILE"
	LevelThirdQuartile  Level = "THIRD_QUARTILE"
	LevelFourthQuartile Level = "FOURTH_QUARTILE"
)

type Profile struct {
	Login     string
	Name      string
	Bio       string
	URL       string
	Followers int
	Following int
}

type Day struct {
	Date      time.Time
	Count     int
	Color     string
	Level     Level
	WeekIndex int
	Row       int
}

type Month struct {
	Name       string
	Year       int
	FirstDay   time.Time
	TotalWeeks int
	StartWeek  int
}

type Summary struct {
	ActiveDays          int
	LongestStreak       int
	CurrentStreak       int
	AveragePerActiveDay float64
	AveragePerWeek      float64
	BestDay             *Day
}

type CalendarInput struct {
	Profile             Profile
	Year                int
	AvailableYears      []int
	TotalContributions  int
	HasAnyContributions bool
	Colors              []string
	IsHalloween         bool
	Months              []Month
	Weeks               [][]*Day
	StartedAt           time.Time
	EndedAt             time.Time
	Now                 time.Time
}

type Calendar struct {
	Profile             Profile
	Year                int
	AvailableYears      []int
	TotalContributions  int
	HasAnyContributions bool
	Colors              []string
	IsHalloween         bool
	Months              []Month
	Weeks               [][]*Day
	Days                []*Day
	StartedAt           time.Time
	EndedAt             time.Time
	Summary             Summary

	dateIndex   map[string]*Day
	position    map[string]int
	weekTotals  []int
	monthTotals map[string]int
}

const dateLayout = "2006-01-02"
