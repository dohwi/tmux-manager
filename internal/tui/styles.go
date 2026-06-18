package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha palette.
// https://github.com/catppuccin/catppuccin
//
//nolint:unused
var (
	ctpBase     = lipgloss.Color("#1e1e2e")
	ctpText     = lipgloss.Color("#cdd6f4")
	ctpSubtext  = lipgloss.Color("#a6adc8")
	ctpOverlay  = lipgloss.Color("#6c7086")
	ctpSurface  = lipgloss.Color("#45475a")
	ctpBlue     = lipgloss.Color("#89b4fa")
	ctpLavender = lipgloss.Color("#b4befe")
	ctpGreen    = lipgloss.Color("#a6e3a1")
	ctpYellow   = lipgloss.Color("#f9e2af")
	ctpPeach    = lipgloss.Color("#fab387")
	ctpPink     = lipgloss.Color("#f5c2e7")
	ctpMaroon   = lipgloss.Color("#eba0ac")
	ctpRed      = lipgloss.Color("#f38ba8")
	ctpMauve    = lipgloss.Color("#cba6f7")
	ctpTeal     = lipgloss.Color("#94e2d5")
	ctpSky      = lipgloss.Color("#89dceb")
)

// Semantic aliases used by the TUI.
var (
	current = ctpSurface
	fg      = ctpText
	comment = ctpOverlay
	cyan    = ctpTeal
	green   = ctpGreen
	pink    = ctpPink
	purple  = ctpLavender
	red     = ctpRed
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

	cfgTag = lipgloss.NewStyle().
		Foreground(purple).
		Render("cfg")

	updateNoticeStyle = lipgloss.NewStyle().
				Foreground(green).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(pink).
				Padding(0, 1)
)

func newDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(pink).
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
