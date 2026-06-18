package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var execCommand = exec.Command

var currentSessionFn = CurrentSession

func IsAvailable() bool {
	if isAvailableFn != nil {
		return isAvailableFn()
	}
	_, err := exec.LookPath("tmux")
	return err == nil
}

func IsInsideTmux() bool {
	if os.Getenv("TMUX") != "" {
		return true
	}

	tty, err := os.Readlink("/proc/self/fd/0")
	if err != nil {
		return false
	}

	cmd := execCommand("tmux", "list-clients", "-F", "#{client_tty}")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	for _, clientTTY := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if clientTTY == tty {
			return true
		}
	}
	return false
}

func ListSessions() ([]Session, error) {
	cmd := execCommand("tmux", "list-sessions", "-F", "#{session_name}:#{session_windows}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tmux list-sessions failed: %w", err)
	}

	current := currentSessionFn()

	var sessions []Session
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		s, ok := parseSessionLine(line, current)
		if !ok {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func parseSessionLine(line, current string) (Session, bool) {
	if line == "" {
		return Session{}, false
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return Session{}, false
	}
	windows, _ := strconv.Atoi(parts[1])
	return Session{
		Name:     parts[0],
		Windows:  windows,
		Attached: parts[0] == current,
	}, true
}

func CurrentSession() string {
	if os.Getenv("TMUX") == "" {
		return ""
	}
	cmd := execCommand("tmux", "display-message", "-p", "#{client_session}")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func SwitchSession(name string) error {
	cmd := execCommand("tmux", "switch-client", "-t", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func NewDetached(name string) error {
	return execCommand("tmux", "new-session", "-d", "-s", name, "-n", "main").Run()
}

func Kill(name string) error {
	return execCommand("tmux", "kill-session", "-t", name).Run()
}

func RenameSession(oldName, newName string) error {
	return execCommand("tmux", "rename-session", "-t", oldName, newName).Run()
}

func HasSession(name string) bool {
	return execCommand("tmux", "has-session", "-t", name).Run() == nil
}
