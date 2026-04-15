package contrib

import "time"

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
