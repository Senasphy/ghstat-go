package ui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	app        lipgloss.Style
	title      lipgloss.Style
	subtitle   lipgloss.Style
	muted      lipgloss.Style
	panel      lipgloss.Style
	panelTitle lipgloss.Style
	accent     lipgloss.Style
	chip       lipgloss.Style
	chipActive lipgloss.Style
	warning    lipgloss.Style
	error      lipgloss.Style
	help       lipgloss.Style
	value      lipgloss.Style
}

func newStyles() styles {
	border := lipgloss.AdaptiveColor{Light: "#B8C2C8", Dark: "#405159"}
	panelTitle := lipgloss.AdaptiveColor{Light: "#12252C", Dark: "#D5E9EE"}
	text := lipgloss.AdaptiveColor{Light: "#142128", Dark: "#E7EFF2"}
	muted := lipgloss.AdaptiveColor{Light: "#5E7078", Dark: "#91A4AD"}
	accent := lipgloss.AdaptiveColor{Light: "#0E7490", Dark: "#7DD3E6"}
	chipBg := lipgloss.AdaptiveColor{Light: "#EEF4F6", Dark: "#182329"}
	activeBg := lipgloss.AdaptiveColor{Light: "#0E7490", Dark: "#1E5B69"}
	activeFg := lipgloss.AdaptiveColor{Light: "#F4FDFF", Dark: "#E9FAFF"}
	titleBg := lipgloss.AdaptiveColor{Light: "#E8F4F7", Dark: "#15252B"}
	error := lipgloss.AdaptiveColor{Light: "#A61B1B", Dark: "#FFB4A9"}
	warning := lipgloss.AdaptiveColor{Light: "#8B5A00", Dark: "#F4C97A"}

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(text).
			Background(titleBg).
			Padding(0, 1).
			Bold(true),
		subtitle: lipgloss.NewStyle().
			Foreground(text),
		muted: lipgloss.NewStyle().
			Foreground(muted),
		panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(border).
			Padding(0, 1),
		panelTitle: lipgloss.NewStyle().
			Foreground(panelTitle).
			Underline(true).
			Bold(true),
		accent: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		chip: lipgloss.NewStyle().
			Foreground(text).
			Background(chipBg).
			Padding(0, 1),
		chipActive: lipgloss.NewStyle().
			Foreground(activeFg).
			Background(activeBg).
			Bold(true).
			Padding(0, 1),
		warning: lipgloss.NewStyle().
			Foreground(warning).
			Bold(true),
		error: lipgloss.NewStyle().
			Foreground(error).
			Bold(true),
		help: lipgloss.NewStyle().
			Foreground(muted),
		value: lipgloss.NewStyle().
			Foreground(text).
			Bold(true),
	}
}
