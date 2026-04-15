package contrib

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
