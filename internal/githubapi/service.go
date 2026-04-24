package githubapi

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"

	"github.com/senasphy/ghstat-go/internal/contrib"
)

const dateLayout = "2006-01-02"

type Service struct {
	client graphql.Client
	now    func() time.Time
}

func NewService(token string) (*Service, error) {
	if strings.TrimSpace(token) == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}

	return &Service{
		client: client,
		now:    time.Now,
	}, nil
}

func (s *Service) LoadCalendar(ctx context.Context, login string, year int) (*contrib.Calendar, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, fmt.Errorf("GitHub username cannot be empty")
	}

	now := s.now().UTC()
	if year == 0 {
		year = now.Year()
	}
	if year > now.Year() {
		return nil, fmt.Errorf("year %d is in the future", year)
	}

	from, to := trailingYearBounds(year, now)
	resp, err := GetContributionCalendar(ctx, s.client, login, from.Format(time.RFC3339), to.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query GitHub API: %w", err)
	}
	if resp.User.Login == "" {
		return nil, fmt.Errorf("GitHub user %q was not found", login)
	}

	return buildCalendar(resp, year, now)
}

func trailingYearBounds(year int, now time.Time) (time.Time, time.Time) {
	day := min(now.Day(), daysInMonth(year, now.Month()))
	end := time.Date(
		year,
		now.Month(),
		day,
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		time.UTC,
	)

	start := end.AddDate(-1, 0, 0)
	return start, end
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func buildCalendar(resp *GetContributionCalendarResponse, year int, now time.Time) (*contrib.Calendar, error) {
	raw := resp.User.ContributionsCollection
	profile := contrib.Profile{
		Login:     resp.User.Login,
		Name:      resp.User.Name,
		Bio:       resp.User.Bio,
		URL:       resp.User.Url,
		Followers: resp.User.Followers.TotalCount,
		Following: resp.User.Following.TotalCount,
	}

	weeks := make([][]*contrib.Day, len(raw.ContributionCalendar.Weeks))
	ordered := make([]*contrib.Day, 0, len(raw.ContributionCalendar.Weeks)*7)

	for weekIndex, rawWeek := range raw.ContributionCalendar.Weeks {
		gridWeek := make([]*contrib.Day, 7)
		for _, rawDay := range rawWeek.ContributionDays {
			date, err := time.Parse(dateLayout, rawDay.Date)
			if err != nil {
				return nil, fmt.Errorf("parse date %q: %w", rawDay.Date, err)
			}

			day := &contrib.Day{
				Date:      date,
				Count:     rawDay.ContributionCount,
				Color:     rawDay.Color,
				Level:     contrib.Level(rawDay.ContributionLevel),
				WeekIndex: weekIndex,
				Row:       rowFromWeekday(rawDay.Weekday),
			}

			gridWeek[day.Row] = day
			ordered = append(ordered, day)
		}
		weeks[weekIndex] = gridWeek
	}

	months := make([]contrib.Month, 0, len(raw.ContributionCalendar.Months))
	for _, rawMonth := range raw.ContributionCalendar.Months {
		firstDay, err := time.Parse(dateLayout, rawMonth.FirstDay)
		if err != nil {
			return nil, fmt.Errorf("parse month %q: %w", rawMonth.FirstDay, err)
		}

		months = append(months, contrib.Month{
			Name:       rawMonth.Name,
			Year:       rawMonth.Year,
			FirstDay:   firstDay,
			TotalWeeks: rawMonth.TotalWeeks,
		})
	}

	years := append([]int(nil), raw.ContributionYears...)
	if !containsYear(years, year) {
		years = append(years, year)
	}
	sort.Ints(years)

	startedAt, err := time.Parse(time.RFC3339, raw.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("parse range start %q: %w", raw.StartedAt, err)
	}
	endedAt, err := time.Parse(time.RFC3339, raw.EndedAt)
	if err != nil {
		return nil, fmt.Errorf("parse range end %q: %w", raw.EndedAt, err)
	}

	return contrib.NewCalendar(contrib.CalendarInput{
		Profile:             profile,
		Year:                year,
		AvailableYears:      years,
		TotalContributions:  raw.ContributionCalendar.TotalContributions,
		HasAnyContributions: raw.HasAnyContributions,
		Colors:              append([]string(nil), raw.ContributionCalendar.Colors...),
		IsHalloween:         raw.ContributionCalendar.IsHalloween,
		Months:              months,
		Weeks:               weeks,
		StartedAt:           startedAt,
		EndedAt:             endedAt,
		Now:                 now,
	})
}

func containsYear(years []int, year int) bool {
	doesContain := slices.Contains(years, year)
	return doesContain
}

func rowFromWeekday(weekday int) int {
	if weekday == 7 {
		return 0
	}
	return weekday
}
