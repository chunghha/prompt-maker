package components

import "github.com/charmbracelet/lipgloss"

const (
	horizontalPadding     = 2
	headerPadding         = 1
	listHorizontalPadding = 2
)

type Styles struct {
	Header, AppName, AppVersion, ModelName, MainContent, Input, StatusBar, StatusText,
	ResubmitHelp, SelectedListItem, ListItem, Spinner, Error lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Header:           lipgloss.NewStyle().Padding(0, headerPadding),
		AppName:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")),
		AppVersion:       lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		ModelName:        lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		MainContent:      lipgloss.NewStyle().Padding(0, horizontalPadding),
		Input:            lipgloss.NewStyle().Padding(1, horizontalPadding),
		StatusBar:        lipgloss.NewStyle().Padding(0, horizontalPadding),
		StatusText:       lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		ResubmitHelp:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")),
		SelectedListItem: lipgloss.NewStyle().Padding(0, 0, 0, listHorizontalPadding).Foreground(lipgloss.Color("208")),
		ListItem:         lipgloss.NewStyle().Padding(0, 0, 0, listHorizontalPadding),
		Spinner:          lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		Error:            lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}
