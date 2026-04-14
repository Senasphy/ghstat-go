package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"ghstat-go/internal/contrib"
	"ghstat-go/internal/githubapi"
)

type Model struct {
	service     *githubapi.Service
	login       string
	initialYear int
	width       int
	height      int
	loading     bool
	loadingYear int
	pendingYear int
	err         error
	calendar    *contrib.Calendar
	selected    *contrib.Day
	anchorMonth time.Month
	anchorDay   int
	spinner     spinner.Model
	help        help.Model
	keys        keyMap
	styles      styles
	now         func() time.Time
}

type calendarLoadedMsg struct {
	calendar *contrib.Calendar
}

type loadFailedMsg struct {
	err error
}

func NewModel(service *githubapi.Service, login string, year int) Model {
	sty := newStyles()
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = sty.accent

	h := help.New()
	h.ShowAll = false

	return Model{
		service:     service,
		login:       login,
		initialYear: year,
		loading:     true,
		loadingYear: year,
		spinner:     s,
		help:        h,
		keys:        defaultKeyMap(),
		styles:      sty,
		now:         time.Now,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("ghstat"),
		m.spinner.Tick,
		m.loadCalendarCmd(m.initialYear),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = max(20, msg.Width-4)
		return m, nil

	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case calendarLoadedMsg:
		m.loading = false
		m.err = nil
		m.calendar = msg.calendar
		m.pendingYear = msg.calendar.Year
		m.selected = m.chooseSelection(msg.calendar)
		return m, nil

	case loadFailedMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		if m.loading || m.calendar == nil {
			return m, nil
		}
		if m.selected == nil {
			m.selected = m.calendar.LastDay()
			if m.selected == nil {
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			m.selected = m.calendar.Move(m.selected, 0, -1)
		case key.Matches(msg, m.keys.Down):
			m.selected = m.calendar.Move(m.selected, 0, 1)
		case key.Matches(msg, m.keys.Left):
			m.selected = m.calendar.Move(m.selected, -1, 0)
		case key.Matches(msg, m.keys.Right):
			m.selected = m.calendar.Move(m.selected, 1, 0)
		case key.Matches(msg, m.keys.FirstDay):
			m.selected = m.calendar.FirstDay()
		case key.Matches(msg, m.keys.LastDay):
			m.selected = m.calendar.LastDay()
		case key.Matches(msg, m.keys.WeekStart):
			m.selected = m.calendar.RowStart(m.selected.Row)
		case key.Matches(msg, m.keys.WeekEnd):
			m.selected = m.calendar.RowEnd(m.selected.Row)
		case key.Matches(msg, m.keys.PrevMonth):
			m.selected = m.calendar.JumpMonth(m.selected, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.selected = m.calendar.JumpMonth(m.selected, 1)
		case key.Matches(msg, m.keys.Today):
			if day := m.calendar.Today(m.now()); day != nil {
				m.selected = day
			}
		case key.Matches(msg, m.keys.PrevYear):
			m.shiftPendingYear(-1)
		case key.Matches(msg, m.keys.NextYear):
			m.shiftPendingYear(1)
		case key.Matches(msg, m.keys.ApplyYear):
			return m.applyPendingYear()
		}

		return m, nil
	}

	return m, nil
}

func (m Model) loadCalendarCmd(year int) tea.Cmd {
	login := m.login
	service := m.service

	return func() tea.Msg {
		calendar, err := service.LoadCalendar(context.Background(), login, year)
		if err != nil {
			return loadFailedMsg{err: err}
		}
		return calendarLoadedMsg{calendar: calendar}
	}
}

func (m *Model) shiftPendingYear(delta int) {
	if m.calendar == nil {
		return
	}

	index := -1
	for i, year := range m.calendar.AvailableYears {
		if year == m.pendingYear {
			index = i
			break
		}
	}
	if index == -1 {
		for i, year := range m.calendar.AvailableYears {
			if year == m.calendar.Year {
				index = i
				break
			}
		}
	}
	if index == -1 {
		return
	}

	target := index + delta
	if target < 0 || target >= len(m.calendar.AvailableYears) {
		return
	}

	m.pendingYear = m.calendar.AvailableYears[target]
}

func (m Model) applyPendingYear() (tea.Model, tea.Cmd) {
	if m.calendar == nil {
		return m, nil
	}
	if m.pendingYear == 0 || m.pendingYear == m.calendar.Year {
		return m, nil
	}

	m.loading = true
	m.loadingYear = m.pendingYear
	m.err = nil
	if m.selected != nil {
		m.anchorMonth = m.selected.Date.Month()
		m.anchorDay = m.selected.Date.Day()
	}

	return m, tea.Batch(m.spinner.Tick, m.loadCalendarCmd(m.loadingYear))
}

func (m Model) chooseSelection(calendar *contrib.Calendar) *contrib.Day {
	if calendar == nil {
		return nil
	}

	if m.anchorMonth != 0 {
		if candidate := calendar.MatchMonthDay(m.anchorMonth, m.anchorDay); candidate != nil {
			m.anchorMonth = 0
			m.anchorDay = 0
			return candidate
		}
	}

	now := m.now().UTC()
	if calendar.Year == now.Year() {
		if today := calendar.Today(now); today != nil {
			return today
		}
	}

	return calendar.LastDay()
}
