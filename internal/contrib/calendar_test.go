package contrib

import (
	"testing"
	"time"
)

func TestCalendarSummaryAndTotals(t *testing.T) {
	calendar := testCalendar(t, "2024-01-01", "2024-01-14", map[string]int{
		"2024-01-02": 2,
		"2024-01-03": 1,
		"2024-01-04": 4,
		"2024-01-10": 1,
		"2024-01-12": 3,
	})

	if got, want := calendar.TotalContributions, 11; got != want {
		t.Fatalf("total contributions = %d, want %d", got, want)
	}
	if got, want := calendar.Summary.ActiveDays, 5; got != want {
		t.Fatalf("active days = %d, want %d", got, want)
	}
	if got, want := calendar.Summary.LongestStreak, 3; got != want {
		t.Fatalf("longest streak = %d, want %d", got, want)
	}
	if got, want := calendar.Summary.CurrentStreak, 0; got != want {
		t.Fatalf("current streak = %d, want %d", got, want)
	}

	best := calendar.Summary.BestDay
	if best == nil || best.Date.Format(dateLayout) != "2024-01-04" || best.Count != 4 {
		t.Fatalf("best day = %#v, want Jan 4 with 4 contributions", best)
	}

	selected := calendar.DayByDate(parseDate(t, "2024-01-10"))
	if got, want := calendar.WeekTotal(selected), 4; got != want {
		t.Fatalf("week total = %d, want %d", got, want)
	}
	if got, want := calendar.MonthTotal(selected), 11; got != want {
		t.Fatalf("month total = %d, want %d", got, want)
	}
	if got, want := calendar.StreakEndingOn(selected), 1; got != want {
		t.Fatalf("streak ending on Jan 10 = %d, want %d", got, want)
	}
}

func TestCalendarMovement(t *testing.T) {
	calendar := testCalendar(t, "2024-01-01", "2024-01-14", map[string]int{
		"2024-01-03": 1,
		"2024-01-10": 2,
		"2024-01-11": 3,
	})

	start := calendar.DayByDate(parseDate(t, "2024-01-03"))
	if start == nil {
		t.Fatal("missing start day")
	}

	if got := calendar.Move(start, 1, 0); got == nil || got.Date.Format(dateLayout) != "2024-01-10" {
		t.Fatalf("move right landed on %v, want Jan 10", got)
	}
	if got := calendar.Move(start, 0, 1); got == nil || got.Date.Format(dateLayout) != "2024-01-04" {
		t.Fatalf("move down landed on %v, want Jan 4", got)
	}
	if got := calendar.RowStart(start.Row); got == nil || got.Date.Format(dateLayout) != "2024-01-03" {
		t.Fatalf("row start landed on %v, want Jan 3", got)
	}
	if got := calendar.RowEnd(start.Row); got == nil || got.Date.Format(dateLayout) != "2024-01-10" {
		t.Fatalf("row end landed on %v, want Jan 10", got)
	}
}

func TestMatchMonthDayClampsToLastDay(t *testing.T) {
	calendar := testCalendar(t, "2024-01-01", "2024-03-31", nil)

	if got := calendar.MatchMonthDay(time.February, 31); got == nil || got.Date.Format(dateLayout) != "2024-02-29" {
		t.Fatalf("match month day landed on %v, want Feb 29", got)
	}

	start := calendar.DayByDate(parseDate(t, "2024-01-31"))
	if got := calendar.JumpMonth(start, 1); got == nil || got.Date.Format(dateLayout) != "2024-02-29" {
		t.Fatalf("jump month landed on %v, want Feb 29", got)
	}
}

func testCalendar(t *testing.T, start, end string, counts map[string]int) *Calendar {
	t.Helper()

	startDate := parseDate(t, start)
	endDate := parseDate(t, end)
	if counts == nil {
		counts = map[string]int{}
	}

	weeks := [][]*Day{}
	currentWeek := make([]*Day, 7)
	weekIndex := 0
	total := 0

	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		if day != startDate && day.Weekday() == time.Sunday {
			weeks = append(weeks, currentWeek)
			currentWeek = make([]*Day, 7)
			weekIndex++
		}

		key := day.Format(dateLayout)
		count := counts[key]
		total += count
		row := weekdayRow(day.Weekday())
		currentWeek[row] = &Day{
			Date:      day,
			Count:     count,
			Color:     "#2da44e",
			Level:     levelForCount(count),
			WeekIndex: weekIndex,
			Row:       row,
		}
	}
	weeks = append(weeks, currentWeek)

	months := []Month{}
	for month := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC); !month.After(endDate); month = month.AddDate(0, 1, 0) {
		months = append(months, Month{
			Name:     month.Month().String(),
			Year:     month.Year(),
			FirstDay: month,
		})
	}

	calendar, err := NewCalendar(CalendarInput{
		Profile:             Profile{Login: "octocat"},
		Year:                startDate.Year(),
		AvailableYears:      []int{startDate.Year()},
		TotalContributions:  total,
		HasAnyContributions: total > 0,
		Colors:              []string{"#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"},
		Months:              months,
		Weeks:               weeks,
		StartedAt:           startDate,
		EndedAt:             endDate,
		Now:                 endDate,
	})
	if err != nil {
		t.Fatalf("new calendar: %v", err)
	}

	return calendar
}

func parseDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}

func weekdayRow(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return 0
	}
	return int(weekday)
}

func levelForCount(count int) Level {
	switch {
	case count <= 0:
		return LevelNone
	case count == 1:
		return LevelFirstQuartile
	case count == 2:
		return LevelSecondQuartile
	case count == 3:
		return LevelThirdQuartile
	default:
		return LevelFourthQuartile
	}
}
