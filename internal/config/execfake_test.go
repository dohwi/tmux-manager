package config

import (
	"os/exec"
	"sync"
	"testing"
)

func captureExec(t *testing.T) (createdNames *[]string, restore func()) {
	t.Helper()
	created := &[]string{}
	mu := &sync.Mutex{}
	orig := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "tmux" && len(args) >= 4 && args[0] == "new-session" {
			// args: new-session -d -s name -P -F "#{window_index}.#{pane_index}" [-n name]
			// Session name is at index 3 (after -d -s).
			if len(args) > 3 && args[1] == "-d" && args[2] == "-s" {
				mu.Lock()
				*created = append(*created, args[3])
				mu.Unlock()
			}
			// Return a valid window.pane index so send-keys succeeds.
			return exec.Command("printf", "1.1")
		}
		// Let other tmux commands (send-keys, split-window, etc.) succeed.
		if name == "tmux" && len(args) > 0 && args[0] == "split-window" {
			return exec.Command("printf", "2")
		}
		if name == "tmux" && len(args) > 0 && args[0] == "new-window" {
			return exec.Command("printf", "1.1")
		}
		return exec.Command("/bin/true")
	}
	return created, func() { execCommand = orig }
}

func failOnCmd(t *testing.T, subcmd string) func() {
	t.Helper()
	orig := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "tmux" && len(args) > 0 && args[0] == subcmd {
			return exec.Command("/bin/false")
		}
		return exec.Command("/bin/true")
	}
	return func() { execCommand = orig }
}
