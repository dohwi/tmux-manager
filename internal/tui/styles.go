package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	bg      = lipgloss.Color("#282a36")
	current = lipgloss.Color("#44475a")
	fg      = lipgloss.Color("#f8f8f2")
	comment = lipgloss.Color("#6272a4")
	cyan    = lipgloss.Color("#8be9fd")
	green   = lipgloss.Color("#50fa7b")
	pink    = lipgloss.Color("#ff79c6")
	purple  = lipgloss.Color("#bd93f9")
	red     = lipgloss.Color("#ff5555")
)

var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 3)

	headerTopBar = lipgloss.NewStyle().
			Foreground(purple)

	headerTextStyle = lipgloss.NewStyle().
			Foreground(fg).
			Bold(true).
			Padding(0, 1)

	sessionCountStyle = lipgloss.NewStyle().
				Foreground(comment).
				Padding(0, 1)

	statusAttached = lipgloss.NewStyle().
			Foreground(green).
			Render("attached")

	statusDetached = lipgloss.NewStyle().
			Foreground(comment).
			Render("detached")

	helpSeparator = lipgloss.NewStyle().
			Foreground(current)

	helpStyle = lipgloss.NewStyle().
			Foreground(comment).
			Padding(0, 1)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Padding(0, 1)

	emptyStyle = lipgloss.NewStyle().
			Foreground(comment)

	emptyHintStyle = lipgloss.NewStyle().
			Foreground(current)
)

func newDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(fg).
		BorderLeftForeground(pink)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(purple)
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(fg)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(comment)
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		Foreground(comment)
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(current)
	return d
}

func newStyledTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "session name"
	ti.CharLimit = 50
	ti.Width = 24
	ti.PromptStyle = lipgloss.NewStyle().Foreground(pink)
	ti.TextStyle = lipgloss.NewStyle().Foreground(fg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(comment)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(cyan)
	return ti
}
