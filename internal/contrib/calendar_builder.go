package contrib

import (
	"fmt"
	"sort"
	"time"
)

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

	summary := buildSummary(days, input.TotalContributions, len(input.Weeks))

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

func buildSummary(days []*Day, totalContributions, weekCount int) Summary {
	summary := Summary{
		AveragePerWeek: float64(totalContributions),
	}

	if weekCount > 0 {
		summary.AveragePerWeek = float64(totalContributions) / float64(weekCount)
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
		summary.AveragePerActiveDay = float64(totalContributions) / float64(summary.ActiveDays)
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

	return summary
}

func monthKey(date time.Time) string {
	return date.Format("2006-01")
}

func firstDayInMonth(days []*Day, year int, month time.Month) *Day {
	for _, day := range days {
		if day.Date.Year() == year && day.Date.Month() == month {
			return day
		}
	}
	return nil
}
