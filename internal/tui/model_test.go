package tui

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dohwi/tmux-manager/internal/tmux"
)

func newModelForTest(sessions []tmux.Session, managed map[string]bool, avail bool) model {
	ti := newStyledTextInput()

	if !avail {
		return model{
			list:        list.New([]list.Item{}, newDelegate(), 0, 0),
			textInput:   ti,
			statusMsg:   "tmux not found. Install tmux first.",
			statusIsErr: true,
		}
	}

	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		m := false
		if managed != nil {
			m = managed[s.Name]
		}
		items[i] = sessionItem{session: s, managed: m}
	}

	l := list.New(items, newDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Select(0)

	return model{
		list:         l,
		textInput:    ti,
		managedNames: managed,
		width:        80,
		height:       24,
	}
}

func mockTmuxExec(t *testing.T, handler func(cmd string, args []string) *exec.Cmd) func() {
	t.Helper()
	return tmux.OverrideExecCommand(func(name string, args ...string) *exec.Cmd {
		return handler(name, args)
	})
}

func TestHandleBrowseKeyEnter_AttachesSelected(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest([]tmux.Session{{Name: "dev", Windows: 2}}, nil, true)
	m.width = 80
	m.height = 24

	m2, cmd := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || cmd() != tea.Quit() {
		t.Error("expected tea.Quit")
	}
	if m2.(model).attachTarget != "dev" {
		t.Errorf("expected attachTarget=dev, got %q", m2.(model).attachTarget)
	}
}

func TestHandleBrowseKeyEnter_EmptyList(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)

	m2, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m2.(model).attachTarget != "" {
		t.Error("expected no attach target on empty list")
	}
}

func TestHandleBrowseKeyCtrlC_Quits(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)

	_, cmd := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || cmd() != tea.Quit() {
		t.Error("expected tea.Quit")
	}
}

func TestHandleBrowseKeyCtrlN_EntersInputMode(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)

	m2, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	if m2.(model).inputMode != inputCreate {
		t.Errorf("expected inputCreate, got %v", m2.(model).inputMode)
	}
}

func TestHandleBrowseKeyCtrlR_EntersRenameMode(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest([]tmux.Session{{Name: "dev", Windows: 3}}, nil, true)

	m2, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlR})
	if m2.(model).inputMode != inputRename {
		t.Errorf("expected inputRename, got %v", m2.(model).inputMode)
	}
	if m2.(model).renameTarget != "dev" {
		t.Errorf("expected renameTarget=dev, got %q", m2.(model).renameTarget)
	}
}

func TestHandleBrowseKeyCtrlR_NothingSelected(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)

	m2, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlR})
	if m2.(model).inputMode != inputNone {
		t.Error("expected inputNone when nothing selected")
	}
}

func TestHandleBrowseKeyCtrlD_KillsSession(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	var killed string
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		if len(args) >= 3 && args[0] == "kill-session" {
			killed = args[2]
		}
		return exec.Command("/bin/true")
	})()

	defer tmux.OverrideCurrentSession(func() string { return "" })()
	m := newModelForTest([]tmux.Session{{Name: "dev", Windows: 2}}, nil, true)

	m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if killed != "dev" {
		t.Errorf("expected kill-session dev, got %q", killed)
	}
}

func TestHandleBrowseKeyCtrlD_HandlesError(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		return exec.Command("/bin/false")
	})()

	defer tmux.OverrideCurrentSession(func() string { return "" })()
	m := newModelForTest([]tmux.Session{{Name: "dev", Windows: 2}}, nil, true)

	m2, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if !m2.(model).statusIsErr {
		t.Error("expected statusIsErr after failed kill")
	}
}

func TestHandleInputKeyEnter_CreatesSession(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	var created string
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		if len(args) >= 4 && args[0] == "new-session" && args[2] == "-s" {
			created = args[3]
		}
		return exec.Command("/bin/true")
	})()

	m := newModelForTest(nil, nil, true)
	m.inputMode = inputCreate
	m.textInput.SetValue("newsession")

	m2, _ := m.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if created != "newsession" {
		t.Errorf("expected new-session newsession, got %q", created)
	}
	if m2.(model).inputMode != inputNone {
		t.Error("expected inputNone after create")
	}
}

func TestHandleInputKeyEnter_CreateFails(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		return exec.Command("/bin/false")
	})()

	m := newModelForTest(nil, nil, true)
	m.inputMode = inputCreate
	m.textInput.SetValue("bad")

	m2, _ := m.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !m2.(model).statusIsErr {
		t.Error("expected statusIsErr after failed create")
	}
}

func TestHandleInputKeyEnter_RenamesSession(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	var oldName, newName string
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		if len(args) >= 4 && args[0] == "rename-session" && args[1] == "-t" {
			oldName = args[2]
			newName = args[3]
		}
		return exec.Command("/bin/true")
	})()

	m := newModelForTest([]tmux.Session{{Name: "old", Windows: 1}}, nil, true)
	m.inputMode = inputRename
	m.renameTarget = "old"
	m.textInput.SetValue("new")

	m2, _ := m.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if oldName != "old" || newName != "new" {
		t.Errorf("expected rename old→new, got %q→%q", oldName, newName)
	}
	if m2.(model).inputMode != inputNone {
		t.Error("expected inputNone after rename")
	}
}

func TestHandleInputKeyEnter_RenameToSame(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	called := false
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		if args[0] == "rename-session" {
			called = true
		}
		return exec.Command("/bin/true")
	})()

	m := newModelForTest([]tmux.Session{{Name: "same", Windows: 1}}, nil, true)
	m.inputMode = inputRename
	m.renameTarget = "same"
	m.textInput.SetValue("same")

	m.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if called {
		t.Error("should not rename to same name")
	}
}

func TestHandleInputKeyEsc_Cancels(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)
	m.inputMode = inputCreate
	m.textInput.SetValue("partial")

	m2, _ := m.handleInputKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m2.(model).inputMode != inputNone {
		t.Error("expected inputNone after cancel")
	}
}

func TestHandleInputKeyEnter_EmptyName(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	m := newModelForTest(nil, nil, true)
	m.inputMode = inputCreate
	m.textInput.SetValue("")

	m2, _ := m.handleInputKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m2.(model).inputMode != inputCreate {
		t.Error("expected inputCreate preserved for empty name")
	}
}

func TestRefreshSessions_HandlesError(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		return exec.Command("/bin/false")
	})()

	m := newModelForTest([]tmux.Session{{Name: "dev", Windows: 1}}, nil, true)
	m.refreshSessions()
	if !m.statusIsErr {
		t.Error("expected statusIsErr after refresh failure")
	}
}

func TestNewModel_TmuxNotFound(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return false })()
	m := newModel(false)
	if !m.statusIsErr {
		t.Error("expected statusIsErr when tmux not found")
	}
}

func TestNewModel_WithSessions(t *testing.T) {
	defer tmux.OverrideIsAvailable(func() bool { return true })()
	defer mockTmuxExec(t, func(cmd string, args []string) *exec.Cmd {
		if args[0] == "list-sessions" {
			return exec.Command("/bin/echo", "dev:2\ninfra:1")
		}
		return exec.Command("/bin/true")
	})()
	defer tmux.OverrideCurrentSession(func() string { return "dev" })()

	m := newModel(false)
	if m.statusIsErr {
		t.Errorf("expected no error, got %q", m.statusMsg)
	}
	if len(m.list.Items()) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(m.list.Items()))
	}
}

func TestSessionItemTitleAttached(t *testing.T) {
	item := sessionItem{session: tmux.Session{Name: "dev", Attached: true}}
	if !strings.HasPrefix(item.Title(), "◉") {
		t.Errorf("expected attached indicator, got %q", item.Title())
	}
}

func TestSessionItemTitleDetached(t *testing.T) {
	item := sessionItem{session: tmux.Session{Name: "dev", Attached: false}}
	if !strings.HasPrefix(item.Title(), "○") {
		t.Errorf("expected detached indicator, got %q", item.Title())
	}
}

func TestSessionItemDescriptionManaged(t *testing.T) {
	item := sessionItem{session: tmux.Session{Name: "dev"}, managed: true}
	if desc := item.Description(); !strings.Contains(desc, "cfg") {
		t.Errorf("expected cfg tag for managed session, got %q", desc)
	}
}

func TestSessionItemFilterValue(t *testing.T) {
	item := sessionItem{session: tmux.Session{Name: "dev"}}
	if fv := item.FilterValue(); fv != "dev" {
		t.Errorf("expected filter value dev, got %q", fv)
	}
}
