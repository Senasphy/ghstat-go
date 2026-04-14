package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	FirstDay  key.Binding
	LastDay   key.Binding
	WeekStart key.Binding
	WeekEnd   key.Binding
	PrevMonth key.Binding
	NextMonth key.Binding
	PrevYear  key.Binding
	NextYear  key.Binding
	ApplyYear key.Binding
	Today     key.Binding
	Help      key.Binding
	Quit      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Left:      key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous week")),
		Right:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next week")),
		FirstDay:  key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "first day")),
		LastDay:   key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "last day")),
		WeekStart: key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "row start")),
		WeekEnd:   key.NewBinding(key.WithKeys("$"), key.WithHelp("$", "row end")),
		PrevMonth: key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "previous month")),
		NextMonth: key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "next month")),
		PrevYear:  key.NewBinding(key.WithKeys("[", "pgup"), key.WithHelp("[", "select previous window")),
		NextYear:  key.NewBinding(key.WithKeys("]", "pgdown"), key.WithHelp("]", "select next window")),
		ApplyYear: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "load selected window")),
		Today:     key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "today")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:      key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.PrevYear, k.NextYear, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down, k.Today},
		{k.FirstDay, k.LastDay, k.WeekStart, k.WeekEnd},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear, k.ApplyYear},
		{k.Help, k.Quit},
	}
}
