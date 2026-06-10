package tmux

import (
	"os/exec"
	"testing"
)

func TestParseSessionLine(t *testing.T) {
	cases := []struct {
		line    string
		current string
		want    Session
		ok      bool
	}{
		{"dev:3", "dev", Session{Name: "dev", Windows: 3, Attached: true}, true},
		{"dev:3", "other", Session{Name: "dev", Windows: 3, Attached: false}, true},
		{"only-name", "x", Session{}, false},
		{"", "x", Session{}, false},
		{"weird:count", "x", Session{Name: "weird", Windows: 0, Attached: false}, true},
		{"dev:abc", "x", Session{Name: "dev", Windows: 0, Attached: false}, true},
	}
	for _, c := range cases {
		got, ok := parseSessionLine(c.line, c.current)
		if ok != c.ok {
			t.Errorf("parseSessionLine(%q) ok=%v, want %v", c.line, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if got != c.want {
			t.Errorf("parseSessionLine(%q) = %+v, want %+v", c.line, got, c.want)
		}
	}
}

func TestListSessions(t *testing.T) {
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/echo", "dev:2\ninfra:1")
	}
	defer func() { execCommand = origExec }()

	origCur := currentSessionFn
	currentSessionFn = func() string { return "dev" }
	defer func() { currentSessionFn = origCur }()

	sessions, err := ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "dev" || sessions[0].Windows != 2 || !sessions[0].Attached {
		t.Errorf("dev session wrong: %+v", sessions[0])
	}
	if sessions[1].Name != "infra" || sessions[1].Windows != 1 || sessions[1].Attached {
		t.Errorf("infra session wrong: %+v", sessions[1])
	}
}

func TestListSessionsError(t *testing.T) {
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/false")
	}
	defer func() { execCommand = origExec }()

	_, err := ListSessions()
	if err == nil {
		t.Error("expected error when tmux fails")
	}
}

func TestHasSession(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/true")
	}
	if !HasSession("any") {
		t.Error("expected HasSession=true on /bin/true")
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/false")
	}
	if HasSession("any") {
		t.Error("expected HasSession=false on /bin/false")
	}
}

func TestCurrentSession(t *testing.T) {
	t.Setenv("TMUX", "")
	if got := CurrentSession(); got != "" {
		t.Errorf("expected empty outside tmux, got %q", got)
	}

	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/echo", "mysession")
	}
	defer func() { execCommand = origExec }()

	if got := CurrentSession(); got != "mysession" {
		t.Errorf("expected mysession, got %q", got)
	}
}

func TestNewDetachedCommand(t *testing.T) {
	var got []string
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return exec.Command("/bin/true")
	}
	defer func() { execCommand = origExec }()

	if err := NewDetached("dev"); err != nil {
		t.Fatal(err)
	}
	if len(got) < 5 || got[0] != "tmux" || got[1] != "new-session" {
		t.Errorf("unexpected command: %v", got)
	}
}

func TestSwitchSessionErrorIncludesOutput(t *testing.T) {
	origExec := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/bin/sh", "-c", "echo 'cannot find session' >&2; exit 1")
	}
	defer func() { execCommand = origExec }()

	err := SwitchSession("missing")
	if err == nil || err.Error() == "" {
		t.Error("expected error with output")
	}
}
