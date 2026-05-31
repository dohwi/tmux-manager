package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func IsAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func ListSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}:#{session_windows}:#{?session_attached,1,0}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tmux list-sessions failed: %w", err)
	}

	var sessions []Session
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		windows, _ := strconv.Atoi(parts[1])
		sessions = append(sessions, Session{
			Name:     parts[0],
			Windows:  windows,
			Attached: parts[2] == "1",
		})
	}
	return sessions, nil
}

func SwitchSession(name string) error {
	return exec.Command("tmux", "switch-client", "-t", name).Run()
}

func NewDetached(name string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name).Run()
}

func NewDetachedWithDir(name, dir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir).Run()
}

func Kill(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

func RenameSession(oldName, newName string) error {
	return exec.Command("tmux", "rename-session", "-t", oldName, newName).Run()
}

func HasSession(name string) bool {
	err := exec.Command("tmux", "has-session", "-t", name).Run()
	return err == nil
}
