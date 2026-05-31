package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tmux-manager/internal/config"
	"tmux-manager/internal/tmux"
)

type inputMode int

const (
	inputNone inputMode = iota
	inputCreate
	inputRename
)

type sessionItem struct {
	session tmux.Session
	managed bool
}

func (i sessionItem) Title() string {
	if i.session.Attached {
		return fmt.Sprintf("◉  %s", i.session.Name)
	}
	return fmt.Sprintf("○  %s", i.session.Name)
}

func (i sessionItem) Description() string {
	if i.session.Attached {
		if i.managed {
			return statusAttached + "  " + cfgTag
		}
		return statusAttached
	}
	if i.managed {
		return statusDetached + "  " + cfgTag
	}
	return statusDetached
}

func (i sessionItem) FilterValue() string {
	return i.session.Name
}

type model struct {
	list          list.Model
	textInput     textinput.Model
	inputMode     inputMode
	renameTarget  string
	attachTarget  string
	statusMsg     string
	statusIsErr   bool
	managedNames  map[string]bool
	width         int
	height        int
}

func Run() (string, error) {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	final := m.(model)
	if final.statusIsErr && final.attachTarget == "" {
		return "", fmt.Errorf("%s", final.statusMsg)
	}
	return final.attachTarget, nil
}

func newModel() model {
	wd, _ := os.Getwd()
	defaultName := filepath.Base(wd)

	ti := newStyledTextInput()
	ti.SetValue(defaultName)

	if !tmux.IsAvailable() {
		return model{
			list:        list.New([]list.Item{}, newDelegate(), 0, 0),
			textInput:   ti,
			statusMsg:   "tmux not found. Install tmux first.",
			statusIsErr: true,
		}
	}

	sessions, err := tmux.ListSessions()
	managed := loadManagedNames()

	var items []list.Item
	if sessions != nil {
		items = make([]list.Item, len(sessions))
		for i, s := range sessions {
			items[i] = sessionItem{session: s, managed: managed[s.Name]}
		}
	}

	l := list.New(items, newDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	m := model{list: l, textInput: ti, managedNames: loadManagedNames()}
	if err != nil {
		m.statusMsg = err.Error()
		m.statusIsErr = true
	}
	return m
}

func loadManagedNames() map[string]bool {
	cfgs, err := config.LoadAll()
	if err != nil {
		return nil
	}
	names := make(map[string]bool, len(cfgs))
	for name := range cfgs {
		names[name] = true
	}
	return names
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-6, msg.Height-9)
		return m, nil

	case tea.KeyMsg:
		if m.inputMode != inputNone {
			return m.handleInputKey(msg)
		}
		return m.handleBrowseKey(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			return m, nil
		}

		if m.inputMode == inputRename {
			if m.renameTarget != name {
				if err := tmux.RenameSession(m.renameTarget, name); err != nil {
					m.statusMsg = fmt.Sprintf("rename failed: %v", err)
					m.statusIsErr = true
				} else {
					m.statusMsg = fmt.Sprintf("renamed %s → %s", m.renameTarget, name)
					m.statusIsErr = false
				}
			}
		} else {
			if err := tmux.NewDetached(name); err != nil {
				m.statusMsg = fmt.Sprintf("create failed: %v", err)
				m.statusIsErr = true
			} else {
				m.statusMsg = fmt.Sprintf("created session: %s", name)
				m.statusIsErr = false
			}
		}

		m.inputMode = inputNone
		m.renameTarget = ""
		m.textInput.Reset()
		m.refreshSessions()
		return m, nil

	case "esc":
		m.inputMode = inputNone
		m.renameTarget = ""
		m.textInput.Reset()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.statusMsg = ""
	m.statusIsErr = false

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		if m.list.SelectedItem() != nil {
			m.attachTarget = m.list.SelectedItem().(sessionItem).session.Name
			return m, tea.Quit
		}
		return m, nil

	case "ctrl+n":
		wd, _ := os.Getwd()
		m.textInput.SetValue(filepath.Base(wd))
		m.inputMode = inputCreate
		return m, tea.Batch(textinput.Blink, m.textInput.Focus())

	case "ctrl+r":
		if item := m.list.SelectedItem(); item != nil {
			session := item.(sessionItem).session
			m.renameTarget = session.Name
			m.textInput.SetValue(session.Name)
			m.inputMode = inputRename
			return m, tea.Batch(textinput.Blink, m.textInput.Focus())
		}
		return m, nil

	case "ctrl+d":
		if m.list.SelectedItem() != nil {
			name := m.list.SelectedItem().(sessionItem).session.Name
			if err := tmux.Kill(name); err != nil {
				m.statusMsg = fmt.Sprintf("delete failed: %v", err)
				m.statusIsErr = true
			} else {
				m.statusMsg = fmt.Sprintf("killed session: %s", name)
				m.statusIsErr = false
			}
			m.refreshSessions()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) refreshSessions() {
	m.managedNames = loadManagedNames()
	sessions, err := tmux.ListSessions()
	if err != nil {
		m.statusMsg = err.Error()
		m.statusIsErr = true
		m.list.SetItems([]list.Item{})
		return
	}
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = sessionItem{session: s, managed: m.managedNames[s.Name]}
	}
	m.list.SetItems(items)
}

func (m model) View() string {
	if m.width == 0 {
		return "initializing..."
	}

	header := m.headerView()

	listView := m.list.View()
	if len(m.list.Items()) == 0 && m.inputMode == inputNone {
		listView = lipgloss.Place(
			m.width-6, m.height-9,
			lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				emptyStyle.Render("No tmux sessions running"),
				emptyHintStyle.Render("^N to create one"),
			),
		)
	}

	sep := helpSeparator.Render(strings.Repeat("▁", m.width-6))

	bottomBar := m.helpBar()

	var statusBar string
	if m.statusMsg != "" {
		if m.statusIsErr {
			statusBar = errorStyle.Render(m.statusMsg)
		} else {
			statusBar = successStyle.Render(m.statusMsg)
		}
	}

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			listView,
			"",
			sep,
			bottomBar,
			statusBar,
		),
	)
}

func (m model) headerView() string {
	title := headerTextStyle.Render("Tmux Manager")
	count := sessionCountStyle.Render(fmt.Sprintf("%d sessions", len(m.list.Items())))

	gap := m.width - 6 - lipgloss.Width(title) - lipgloss.Width(count)
	if gap < 1 {
		gap = 1
	}
	titleLine := title + strings.Repeat(" ", gap) + count

	topBar := headerTopBar.Render(strings.Repeat("▔", m.width-6))

	return lipgloss.JoinVertical(lipgloss.Left, topBar, titleLine)
}

func (m model) helpBar() string {
	if m.inputMode != inputNone {
		var label string
		if m.inputMode == inputRename {
			label = fmt.Sprintf("Rename (%s):", m.renameTarget)
		} else {
			label = "Create:"
		}

		left := lipgloss.JoinHorizontal(lipgloss.Top,
			inputLabelStyle.Render(label),
			m.textInput.View(),
		)

		hint := helpStyle.Render("⏎ confirm  ⎋ cancel")
		gap := m.width - 6 - lipgloss.Width(left) - lipgloss.Width(hint)
		if gap < 1 {
			gap = 1
		}

		return lipgloss.JoinHorizontal(lipgloss.Left,
			left,
			strings.Repeat(" ", gap),
			hint,
		)
	}

	return helpStyle.Render("↑↓ Navigate  ⏎ Attach  ^N New  ^R Rename  ^D Delete  ^C Quit")
}
