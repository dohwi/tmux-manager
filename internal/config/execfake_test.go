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
			mu.Lock()
			*created = append(*created, args[3])
			mu.Unlock()
			return exec.Command("/bin/true")
		}
		return exec.Command("/bin/false")
	}
	return created, func() { execCommand = orig }
}
