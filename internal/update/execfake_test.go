package update

import (
	"os/exec"
	"testing"
)

func fakeExec(t *testing.T, wantName string, wantArgs ...string) func(string, ...string) *exec.Cmd {
	t.Helper()
	return func(name string, args ...string) *exec.Cmd {
		if name != wantName {
			t.Errorf("exec name = %q, want %q", name, wantName)
		}
		if len(args) != len(wantArgs) {
			t.Errorf("exec args len = %d, want %d (%v)", len(args), len(wantArgs), args)
		} else {
			for i, a := range args {
				if a != wantArgs[i] {
					t.Errorf("exec arg[%d] = %q, want %q", i, a, wantArgs[i])
				}
			}
		}
		return exec.Command("/bin/false")
	}
}
