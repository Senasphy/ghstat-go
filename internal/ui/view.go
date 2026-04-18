package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"ghstat-go/internal/contrib"
)

const (
	dayLabelWidth = 4
	cellWidth     = 2
	cellGap       = 1
	rowGap        = 1
	panelGap      = 3

	minCompactSplitWidth = 92
	minChartWidth        = 42
	minDetailsWidth      = 24
	maxDetailsWidth      = 32
	yearChipWindowSize   = 3
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	contentWidth := m.width - m.styles.app.GetHorizontalFrameSize()
	if contentWidth < 1 {
		contentWidth = 1
	}
	header := m.renderHeader(contentWidth)
	helpText := m.styles.help.Render(m.help.View(m.keys))
	sections := []string{header}

	if m.calendar == nil {
		sections = append(sections, m.renderEmptyState(contentWidth))
	} else {
		main := m.renderMain(contentWidth)
		availableHeight := m.height - m.styles.app.GetVerticalFrameSize() - lipgloss.Height(header) - lipgloss.Height(helpText) - 4
		if availableHeight > lipgloss.Height(main) {
			main = lipgloss.PlaceVertical(availableHeight, lipgloss.Center, main)
		}
		sections = append(sections, main)
	}

	sections = append(sections, helpText)
	return m.styles.app.Width(contentWidth).Render(strings.Join(sections, "\n\n"))
}

func (m Model) renderHeader(width int) string {
	subject := m.login
	if m.calendar != nil {
		subject = m.calendar.Profile.Login
		if m.calendar.Profile.Name != "" {
			subject = fmt.Sprintf("%s (@%s)", m.calendar.Profile.Name, m.calendar.Profile.Login)
		}
	}

	title := lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.styles.title.Render("ghstat"),
		"  ",
		m.styles.subtitle.Render(subject),
	)

	metaParts := []string{}
	if m.calendar != nil {
		metaParts = append(metaParts,
			fmt.Sprintf("%s total", formatNumber(m.calendar.TotalContributions)),
			fmt.Sprintf("%d active days", m.calendar.Summary.ActiveDays),
			fmt.Sprintf("%d-day best streak", m.calendar.Summary.LongestStreak),
			fmt.Sprintf("%d followers", m.calendar.Profile.Followers),
		)
	}

	if len(metaParts) > 0 {
		title = lipgloss.JoinVertical(lipgloss.Left, title, m.styles.muted.Render(strings.Join(metaParts, "  •  ")))
	}

	if m.calendar != nil && m.calendar.Profile.Bio != "" {
		title = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			m.styles.muted.Render(truncate(m.calendar.Profile.Bio, width)),
		)
	}

	if status := m.renderStatus(); status != "" {
		title = lipgloss.JoinVertical(lipgloss.Left, title, status)
	}

	return title
}

func (m Model) renderStatus() string {
	switch {
	case m.loading:
		if m.calendar == nil {
			return ""
		}
		return m.styles.warning.Render(fmt.Sprintf("%s loading %d", m.spinner.View(), m.loadingYearOrCurrent()))
	case m.err != nil:
		return m.styles.error.Render(m.err.Error())
	default:
		return ""
	}
}

func (m Model) renderEmptyState(width int) string {
	body := "No contribution data loaded yet."
	if m.loading {
		title := m.styles.accent.Render(fmt.Sprintf("%s Loading window %d", m.spinner.View(), m.loadingYearOrCurrent()))
		subtitle := m.styles.muted.Render("Fetching your contribution calendar from GitHub")
		hint := m.styles.muted.Render("Press q to quit")
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			"",
			subtitle,
			"",
			hint,
		)
		return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
	}
	if m.err != nil {
		body = m.err.Error()
	}

	panel := m.panelWithTotalWidth(width)
	return panel.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.panelTitle.Render("Loading"),
			body,
		),
	)
}

func (m Model) renderMain(width int) string {
	if width < minCompactSplitWidth {
		return strings.Join([]string{
			m.renderChartPanel(width),
			m.renderDetailsPanel(width),
		}, "\n\n")
	}

	frame := m.styles.panel.GetHorizontalFrameSize()
	detailsWidth := min(maxDetailsWidth, max(minDetailsWidth, width/4))
	availableChartWidth := max(minChartWidth, width-detailsWidth-panelGap)
	chartGridWidth := dayLabelWidth + len(m.calendar.Weeks)*(cellWidth+cellGap)
	if len(m.calendar.Weeks) > 0 {
		chartGridWidth -= cellGap
	}
	chartMaxWidth := max(minChartWidth, chartGridWidth+frame)
	chartWidth := min(availableChartWidth, chartMaxWidth)
	detailsWidth = width - chartWidth - panelGap
	if detailsWidth < minDetailsWidth {
		detailsWidth = minDetailsWidth
		chartWidth = max(minChartWidth, width-detailsWidth-panelGap)
	}

	left := m.renderChartPanel(chartWidth)
	right := m.renderDetailsPanel(detailsWidth)
	panelHeight := max(lipgloss.Height(left), lipgloss.Height(right))
	left = lipgloss.PlaceVertical(panelHeight, lipgloss.Top, left)
	right = lipgloss.PlaceVertical(panelHeight, lipgloss.Top, right)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().MarginRight(panelGap).Render(left),
		right,
	)
}

func (m Model) renderChartPanel(width int) string {
	panel := m.panelWithTotalWidth(width)
	innerWidth := max(24, width-m.styles.panel.GetHorizontalFrameSize())

	content := []string{
		m.styles.panelTitle.Render("Contribution Calendar"),
		"",
	}

	if m.calendar == nil || len(m.calendar.Weeks) == 0 {
		content = append(content, "No chart data available.")
		return panel.Render(strings.Join(content, "\n"))
	}

	if innerWidth < dayLabelWidth+cellWidth {
		content = append(content, "Terminal is too narrow to render the chart.")
		return panel.Render(strings.Join(content, "\n"))
	}

	availableGridWidth := innerWidth - dayLabelWidth
	safeGridWidth := availableGridWidth - (cellWidth + cellGap)
	if safeGridWidth < cellWidth {
		safeGridWidth = cellWidth
	}
	maxWeeks := (safeGridWidth + cellGap) / (cellWidth + cellGap)
	maxWeeks = max(1, min(maxWeeks, len(m.calendar.Weeks)))

	start, end := m.visibleWeeks(maxWeeks)
	// When shrinking is needed, always remove from the left/start side so the
	// last column stays visible and never wraps.
	start = max(0, end-maxWeeks)
	for !m.gridFits(innerWidth, start, end) && start < end-1 {
		start++
	}

	content = append(content, m.renderMonthHeader(start, end))
	content = append(content, m.renderRows(start, end)...)
	content = append(content, "")
	content = append(content, m.renderLegend(innerWidth))

	if end-start < len(m.calendar.Weeks) {
		content = append(content, m.styles.muted.Render(fmt.Sprintf("Weeks %d-%d of %d", start+1, end, len(m.calendar.Weeks))))
	}

	return panel.Render(strings.Join(content, "\n"))
}

func (m Model) renderDetailsPanel(width int) string {
	panel := m.panelWithTotalWidth(width)
	content := []string{m.styles.panelTitle.Render("Selection")}

	if m.calendar == nil || m.selected == nil {
		content = append(content, "No day selected.")
		return panel.Render(strings.Join(content, "\n"))
	}

	day := m.selected
	content = append(content,
		m.styles.value.Render(fmt.Sprintf("%s contributions", formatNumber(day.Count))),
		m.styles.subtitle.Render(day.Date.Format("Mon, Jan 2 2006")),
		m.styles.muted.Render(humanizeLevel(day.Level)),
		"",
		fmt.Sprintf("Week total      %s", formatNumber(m.calendar.WeekTotal(day))),
		fmt.Sprintf("Month total     %s", formatNumber(m.calendar.MonthTotal(day))),
		fmt.Sprintf("Streak here     %d days", m.calendar.StreakEndingOn(day)),
		fmt.Sprintf("Average / week  %s", formatFloat(m.calendar.Summary.AveragePerWeek)),
	)

	if best := m.calendar.Summary.BestDay; best != nil {
		content = append(content,
			"",
			fmt.Sprintf("Best day        %s", best.Date.Format("Jan 2")),
			fmt.Sprintf("Peak count      %s", formatNumber(best.Count)),
			fmt.Sprintf("Current streak  %d days", m.calendar.Summary.CurrentStreak),
		)
	}

	yearsWidth := width - panel.GetHorizontalFrameSize()
	content = append(content, "", "Years", m.renderYearChips(yearsWidth))
	if m.pendingYear != 0 && m.pendingYear != m.calendar.Year {
		content = append(content, m.styles.muted.Render(fmt.Sprintf("Press Enter to load %d", m.pendingYear)))
	}
	return panel.Render(strings.Join(content, "\n"))
}

func (m Model) renderMonthHeader(start, end int) string {
	weeks := end - start
	width := 0
	if weeks > 0 {
		width = weeks*cellWidth + (weeks-1)*cellGap
	}
	runes := []rune(strings.Repeat(" ", width))

	for _, month := range m.calendar.Months {
		if month.StartWeek < start || month.StartWeek >= end {
			continue
		}
		label := []rune(shortMonth(month.Name))
		offset := (month.StartWeek - start) * (cellWidth + cellGap)
		for index, char := range label {
			if offset+index >= len(runes) {
				break
			}
			runes[offset+index] = char
		}
	}

	return strings.Repeat(" ", dayLabelWidth) + string(runes)
}

func (m Model) renderRows(start, end int) []string {
	labels := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	rows := make([]string, 0, 7+(7-1)*rowGap)

	for row := range 7 {
		var builder strings.Builder
		for week := start; week < end; week++ {
			if week > start {
				builder.WriteString(strings.Repeat(" ", cellGap))
			}
			builder.WriteString(m.renderCell(m.calendar.DayAt(week, row)))
		}
		rows = append(rows, fmt.Sprintf("%-*s%s", dayLabelWidth, labels[row], builder.String()))
		if row < 6 {
			for range rowGap {
				rows = append(rows, "")
			}
		}
	}

	return rows
}

func (m Model) renderCell(day *contrib.Day) string {
	style := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)
	if day == nil {
		return style.Render(strings.Repeat(" ", cellWidth))
	}

	if m.selected != nil && day.Date.Equal(m.selected.Date) {
		return style.
			Bold(true).
			Foreground(readableColor(day.Color)).
			Background(lipgloss.Color(day.Color)).
			Render("[]")
	}

	return style.Background(lipgloss.Color(day.Color)).Render(strings.Repeat(" ", cellWidth))
}

func (m Model) renderLegend(width int) string {
	colors := m.calendar.Colors
	if len(colors) == 0 {
		colors = []string{"#9AA0A6", "#7FB77E", "#63A05D", "#2E7D32", "#145A1F"}
	}

	levels := []contrib.Level{
		contrib.LevelNone,
		contrib.LevelFirstQuartile,
		contrib.LevelSecondQuartile,
		contrib.LevelThirdQuartile,
		contrib.LevelFourthQuartile,
	}

	cells := make([]string, 0, len(levels))
	for index, level := range levels {
		day := &contrib.Day{
			Color: colors[min(index, len(colors)-1)],
			Level: level,
		}
		cells = append(cells, m.renderCell(day))
	}

	legend := m.styles.muted.Render("Less") + " " + strings.Join(cells, strings.Repeat(" ", cellGap)) + " " + m.styles.muted.Render("More")
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, legend)
}

func (m Model) renderYearChips(width int) string {
	pending := m.pendingYear
	if pending == 0 {
		pending = m.calendar.Year
	}

	years := m.visibleYearWindow(pending, yearChipWindowSize)
	chips := make([]string, 0, len(years))
	for _, year := range years {
		style := m.styles.chip
		label := strconv.Itoa(year)
		if year == m.calendar.Year && pending != m.calendar.Year {
			label += "•"
		}
		if year == pending {
			style = m.styles.chipActive
		}
		chips = append(chips, style.Render(label))
	}

	line := strings.Join(chips, " ")
	if lipgloss.Width(line) <= width {
		return line
	}

	return strings.Join(chips, "\n")
}

func (m Model) visibleYearWindow(center, limit int) []int {
	years := m.calendar.AvailableYears
	if limit <= 0 || len(years) <= limit {
		return years
	}

	index := 0
	for i, year := range years {
		if year == center {
			index = i
			break
		}
	}

	start := index - limit/2
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(years) {
		end = len(years)
		start = end - limit
	}

	return years[start:end]
}

func (m Model) gridFits(innerWidth, start, end int) bool {
	if lipgloss.Width(m.renderMonthHeader(start, end)) > innerWidth {
		return false
	}
	for _, row := range m.renderRows(start, end) {
		if lipgloss.Width(row) > innerWidth {
			return false
		}
	}
	return true
}

func (m Model) visibleWeeks(maxWeeks int) (int, int) {
	total := len(m.calendar.Weeks)
	if maxWeeks >= total {
		return 0, total
	}

	// No selection: pin to the end (most recent weeks), GitHub-style
	if m.selected == nil {
		return total - maxWeeks, total
	}

	selectedWeek := m.selected.WeekIndex
	start := selectedWeek - maxWeeks/2
	start = max(0, start)

	end := start + maxWeeks
	if end > total {
		end = total
		start = end - maxWeeks
	}

	return start, end
}

func (m Model) loadingYearOrCurrent() int {
	if m.loadingYear != 0 {
		return m.loadingYear
	}
	return time.Now().Year()
}

func (m Model) panelWithTotalWidth(totalWidth int) lipgloss.Style {
	contentWidth := totalWidth - m.styles.panel.GetHorizontalFrameSize()
	if contentWidth < 1 {
		contentWidth = 1
	}
	return m.styles.panel.Width(contentWidth)
}

func humanizeLevel(level contrib.Level) string {
	switch level {
	case contrib.LevelNone:
		return "No contributions"
	case contrib.LevelFirstQuartile:
		return "Low activity"
	case contrib.LevelSecondQuartile:
		return "Steady activity"
	case contrib.LevelThirdQuartile:
		return "High activity"
	case contrib.LevelFourthQuartile:
		return "Peak activity"
	default:
		return "Activity"
	}
}

func shortMonth(name string) string {
	if len(name) < 3 {
		return name
	}
	return name[:3]
}

func readableColor(hex string) lipgloss.Color {
	clean := strings.TrimPrefix(hex, "#")
	if len(clean) != 6 {
		return lipgloss.Color("#111111")
	}

	value, err := strconv.ParseUint(clean, 16, 32)
	if err != nil {
		return lipgloss.Color("#111111")
	}

	r := float64((value >> 16) & 0xff)
	g := float64((value >> 8) & 0xff)
	b := float64(value & 0xff)
	luminance := (0.299*r + 0.587*g + 0.114*b) / 255
	if luminance > 0.6 {
		return lipgloss.Color("#111111")
	}
	return lipgloss.Color("#F8FBF8")
}

func formatFloat(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "0"
	}
	return fmt.Sprintf("%.1f", value)
}

func formatNumber(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	raw := strconv.Itoa(value)
	if len(raw) <= 3 {
		return sign + raw
	}

	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return sign + strings.Join(parts, ",")
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if lipgloss.Width(string(runes)) <= width {
		return value
	}
	if width <= 1 {
		return ""
	}

	trimmed := make([]rune, 0, len(runes))
	for _, r := range runes {
		next := string(append(trimmed, r))
		if lipgloss.Width(next+"…") > width {
			break
		}
		trimmed = append(trimmed, r)
	}

	return string(trimmed) + "…"
}
