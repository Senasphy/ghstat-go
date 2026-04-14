package contrib

import (
	"fmt"
	"sort"
	"time"
)

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

func NewCalendar(input CalendarInput) (*Calendar, error) {
	days := make([]*Day, 0, len(input.Weeks)*7)
	dateIndex := make(map[string]*Day, len(input.Weeks)*7)
	position := make(map[string]int, len(input.Weeks)*7)
	weekTotals := make([]int, len(input.Weeks))
	monthTotals := make(map[string]int, len(input.Months))

	for _, week := range input.Weeks {
		for _, day := range week {
			if day == nil {
				continue
			}
			days = append(days, day)
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.Before(days[j].Date)
	})

	for index, day := range days {
		key := day.Date.Format(dateLayout)
		dateIndex[key] = day
		position[key] = index
		weekTotals[day.WeekIndex] += day.Count
		monthTotals[monthKey(day.Date)] += day.Count
	}

	months := append([]Month(nil), input.Months...)
	for index := range months {
		day := dateIndex[months[index].FirstDay.Format(dateLayout)]
		if day == nil {
			day = firstDayInMonth(days, months[index].Year, months[index].FirstDay.Month())
		}
		if day == nil {
			return nil, fmt.Errorf("month %q does not exist in calendar", months[index].FirstDay.Format(dateLayout))
		}
		months[index].StartWeek = day.WeekIndex
	}

	summary := Summary{
		AveragePerWeek: float64(input.TotalContributions),
	}

	if len(input.Weeks) > 0 {
		summary.AveragePerWeek = float64(input.TotalContributions) / float64(len(input.Weeks))
	}

	for _, day := range days {
		if day.Count > 0 {
			summary.ActiveDays++
			if summary.BestDay == nil || day.Count > summary.BestDay.Count {
				summary.BestDay = day
			}
		}
	}
	if summary.ActiveDays > 0 {
		summary.AveragePerActiveDay = float64(input.TotalContributions) / float64(summary.ActiveDays)
	}

	streak := 0
	for _, day := range days {
		if day.Count > 0 {
			streak++
			if streak > summary.LongestStreak {
				summary.LongestStreak = streak
			}
			continue
		}
		streak = 0
	}

	for index := len(days) - 1; index >= 0; index-- {
		if days[index].Count == 0 {
			break
		}
		summary.CurrentStreak++
	}

	return &Calendar{
		Profile:             input.Profile,
		Year:                input.Year,
		AvailableYears:      append([]int(nil), input.AvailableYears...),
		TotalContributions:  input.TotalContributions,
		HasAnyContributions: input.HasAnyContributions,
		Colors:              append([]string(nil), input.Colors...),
		IsHalloween:         input.IsHalloween,
		Months:              months,
		Weeks:               input.Weeks,
		Days:                days,
		StartedAt:           input.StartedAt,
		EndedAt:             input.EndedAt,
		Summary:             summary,
		dateIndex:           dateIndex,
		position:            position,
		weekTotals:          weekTotals,
		monthTotals:         monthTotals,
	}, nil
}

func (c *Calendar) FirstDay() *Day {
	if len(c.Days) == 0 {
		return nil
	}
	return c.Days[0]
}

func (c *Calendar) LastDay() *Day {
	if len(c.Days) == 0 {
		return nil
	}
	return c.Days[len(c.Days)-1]
}

func (c *Calendar) DayAt(week, row int) *Day {
	if week < 0 || week >= len(c.Weeks) || row < 0 || row >= 7 {
		return nil
	}
	return c.Weeks[week][row]
}

func (c *Calendar) DayByDate(date time.Time) *Day {
	return c.dateIndex[date.Format(dateLayout)]
}

func (c *Calendar) MatchMonthDay(month time.Month, day int) *Day {
	if day < 1 {
		return nil
	}

	if lastDay := daysInMonth(c.Year, month); day > lastDay {
		day = lastDay
	}

	target := time.Date(c.Year, month, day, 0, 0, 0, 0, time.UTC)
	if exact := c.DayByDate(target); exact != nil {
		return exact
	}

	for fallbackDay := day - 1; fallbackDay >= 1; fallbackDay-- {
		target = time.Date(c.Year, month, fallbackDay, 0, 0, 0, 0, time.UTC)
		if candidate := c.DayByDate(target); candidate != nil {
			return candidate
		}
	}

	for _, candidate := range c.Days {
		if candidate.Date.Month() == month {
			return candidate
		}
	}

	return nil
}

func (c *Calendar) Move(day *Day, dx, dy int) *Day {
	if day == nil {
		return c.LastDay()
	}
	if dx == 0 && dy == 0 {
		return day
	}

	targetWeek := day.WeekIndex + dx
	targetRow := day.Row + dy
	if targetRow < 0 || targetRow >= 7 {
		return day
	}

	if targetWeek < 0 {
		targetWeek = 0
	}
	if targetWeek >= len(c.Weeks) {
		targetWeek = len(c.Weeks) - 1
	}

	if candidate := c.DayAt(targetWeek, targetRow); candidate != nil {
		return candidate
	}

	if candidate := c.closestInWeek(targetWeek, targetRow); candidate != nil {
		return candidate
	}

	if dx != 0 {
		step := sign(dx)
		for week := targetWeek + step; week >= 0 && week < len(c.Weeks); week += step {
			if candidate := c.DayAt(week, targetRow); candidate != nil {
				return candidate
			}
			if candidate := c.closestInWeek(week, targetRow); candidate != nil {
				return candidate
			}
		}
	}

	return day
}

func (c *Calendar) JumpMonth(day *Day, delta int) *Day {
	if day == nil || len(c.Months) == 0 || delta == 0 {
		return day
	}

	current := 0
	for index, month := range c.Months {
		if month.Year == day.Date.Year() && month.FirstDay.Month() == day.Date.Month() {
			current = index
			break
		}
	}

	target := current + delta
	if target < 0 {
		target = 0
	}
	if target >= len(c.Months) {
		target = len(c.Months) - 1
	}

	return c.MatchMonthDay(c.Months[target].FirstDay.Month(), day.Date.Day())
}

func (c *Calendar) RowStart(row int) *Day {
	for week := 0; week < len(c.Weeks); week++ {
		if day := c.DayAt(week, row); day != nil {
			return day
		}
	}
	return nil
}

func (c *Calendar) RowEnd(row int) *Day {
	for week := len(c.Weeks) - 1; week >= 0; week-- {
		if day := c.DayAt(week, row); day != nil {
			return day
		}
	}
	return nil
}

func (c *Calendar) Today(now time.Time) *Day {
	return c.DayByDate(now.UTC())
}

func (c *Calendar) Month(day *Day) *Month {
	if day == nil {
		return nil
	}
	for index := range c.Months {
		if c.Months[index].Year == day.Date.Year() && c.Months[index].FirstDay.Month() == day.Date.Month() {
			return &c.Months[index]
		}
	}
	return nil
}

func (c *Calendar) MonthTotal(day *Day) int {
	if day == nil {
		return 0
	}
	return c.monthTotals[monthKey(day.Date)]
}

func (c *Calendar) WeekTotal(day *Day) int {
	if day == nil || day.WeekIndex < 0 || day.WeekIndex >= len(c.weekTotals) {
		return 0
	}
	return c.weekTotals[day.WeekIndex]
}

func (c *Calendar) StreakEndingOn(day *Day) int {
	if day == nil || day.Count == 0 {
		return 0
	}
	index, ok := c.position[day.Date.Format(dateLayout)]
	if !ok {
		return 0
	}

	streak := 0
	for ; index >= 0 && c.Days[index].Count > 0; index-- {
		streak++
	}
	return streak
}

func (c *Calendar) closestInWeek(week, row int) *Day {
	offsets := []int{1, -1, 2, -2, 3, -3, 4, -4, 5, -5, 6, -6}
	for _, offset := range offsets {
		candidate := c.DayAt(week, row+offset)
		if candidate != nil {
			return candidate
		}
	}
	return nil
}

func monthKey(date time.Time) string {
	return date.Format("2006-01")
}

func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func firstDayInMonth(days []*Day, year int, month time.Month) *Day {
	for _, day := range days {
		if day.Date.Year() == year && day.Date.Month() == month {
			return day
		}
	}
	return nil
}

const dateLayout = "2006-01-02"
