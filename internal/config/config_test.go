package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cases := []struct {
		in, want string
	}{
		{"~/projects", filepath.Join(dir, "projects")},
		{"~", "~"},
		{"/abs/path", "/abs/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}
	for _, c := range cases {
		if got := expandHome(c.in); got != c.want {
			t.Errorf("expandHome(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPaneCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cases := []struct {
		name string
		p    PaneConfig
		want string
	}{
		{"cmd only", PaneConfig{Command: "nvim"}, "nvim"},
		{"dir only", PaneConfig{Directory: "~/proj"}, "cd " + filepath.Join(dir, "proj")},
		{"cmd+dir", PaneConfig{Command: "nvim", Directory: "~/proj"}, "cd " + filepath.Join(dir, "proj") + " && nvim"},
		{"empty", PaneConfig{}, ""},
		{"abs dir", PaneConfig{Command: "ls", Directory: "/tmp"}, "cd /tmp && ls"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := paneCommand(c.p); got != c.want {
				t.Errorf("paneCommand = %q, want %q", got, c.want)
			}
		})
	}
}

func TestLoadDirMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	got, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("expected nil error for missing dir, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil map, got %v", got)
	}
}

func TestLoadDirValid(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", `
sessions:
  dev:
    windows:
      - panes:
          - command: nvim
            directory: ~/proj
  db:
    windows:
      - name: main
        panes:
          - command: psql
`)

	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(configs))
	}
	if configs["dev"].Windows[0].Panes[0].Command != "nvim" {
		t.Errorf("dev.windows[0].panes[0].command = %q", configs["dev"].Windows[0].Panes[0].Command)
	}
	if len(configs["db"].Windows) != 1 || configs["db"].Windows[0].Name != "main" {
		t.Errorf("db.windows mismatch: %+v", configs["db"])
	}
}

func TestLoadDirSkipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", `sessions: {x: {}}`)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore me"), 0o644); err != nil {
		t.Fatal(err)
	}
	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := configs["x"]; !ok {
		t.Error("expected session x loaded")
	}
}

func TestLoadDirInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad.yaml", ":\n  - not: [valid")
	_, err := LoadDir(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadDirEmptySessionName(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", `
sessions:
  "":
    windows:
      - panes:
          - command: nvim
`)
	_, err := LoadDir(dir)
	if err == nil || !contains(err.Error(), "name") {
		t.Errorf("expected name error, got %v", err)
	}
}

func TestLoadDirDuplicateName(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", `sessions: {dup: {}}`)
	writeYAML(t, dir, "b.yaml", `sessions: {dup: {}}`)
	_, err := LoadDir(dir)
	if err == nil || !contains(err.Error(), "duplicate") {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestLoadDirSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeYAML(t, dir, "a.yaml", `sessions: {x: {}}`)
	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 {
		t.Errorf("expected 1 session (dir skipped), got %d", len(configs))
	}
}

func TestRestoreAllSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, ".config", "tmux-manager", "sessions")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeYAML(t, cfgDir, "a.yaml", `sessions: {existing: {}, fresh: {}}`)

	origHas := hasSessionFn
	hasSessionFn = func(name string) bool { return name == "existing" }
	defer func() { hasSessionFn = origHas }()

	created, restoreExec := captureExec(t)
	defer restoreExec()

	if err := RestoreAll(); err != nil {
		t.Fatal(err)
	}
	if len(*created) != 1 || (*created)[0] != "fresh" {
		t.Errorf("expected only 'fresh' created, got %v", *created)
	}
}

func TestRestoreAllEmptyConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, ".config", "tmux-manager", "sessions")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	if err := RestoreAll(); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "no sessions defined") {
		t.Errorf("expected hint message, got %q", buf.String())
	}
}

func TestCreateSessionSplitWindowError(t *testing.T) {
	defer failOnCmd(t, "split-window")()
	cfg := SessionConfig{Windows: []WindowConfig{{Panes: []PaneConfig{{Command: "nvim"}, {Command: "ls", Direction: "right"}}}}}
	err := createSession("test", cfg)
	if err == nil || !contains(err.Error(), "split-window") {
		t.Errorf("expected split-window error, got: %v", err)
	}
}

func TestCreateSessionSendKeysError(t *testing.T) {
	defer failOnCmd(t, "send-keys")()
	cfg := SessionConfig{Windows: []WindowConfig{{Panes: []PaneConfig{{Command: "nvim"}}}}}
	err := createSession("test", cfg)
	if err == nil || !contains(err.Error(), "send-keys") {
		t.Errorf("expected send-keys error, got: %v", err)
	}
}

func TestCreateSessionNewWindowError(t *testing.T) {
	defer failOnCmd(t, "new-window")()
	cfg := SessionConfig{
		Windows: []WindowConfig{
			{Name: "win1"},
			{Name: "win2"},
		},
	}
	err := createSession("test", cfg)
	if err == nil || !contains(err.Error(), "new-window") {
		t.Errorf("expected new-window error, got: %v", err)
	}
}

func TestCreateSessionSelectPaneError(t *testing.T) {
	defer failOnCmd(t, "select-pane")()
	cfg := SessionConfig{Windows: []WindowConfig{{Panes: []PaneConfig{{Name: "title", Command: "nvim"}}}}}
	err := createSession("test", cfg)
	if err == nil || !contains(err.Error(), "select-pane") {
		t.Errorf("expected select-pane error, got: %v", err)
	}
}

func TestValidateCommandAcceptsSafe(t *testing.T) {
	safe := []string{"nvim", "npm run dev", "cd ~/proj && ls", "echo hi", "", "htop", "docker compose up"}
	for _, s := range safe {
		t.Run(s, func(t *testing.T) {
			if err := validateCommand(s); err != nil {
				t.Errorf("expected valid, got err: %v", err)
			}
		})
	}
}

func TestValidateCommandRejectsMetachars(t *testing.T) {
	dangerous := []string{
		"rm -rf $HOME",
		"ls; whoami",
		"`backtick`",
		"echo > /tmp/out",
		"cat < /etc/passwd",
		"cmd|pipe",
	}
	for _, s := range dangerous {
		t.Run(s, func(t *testing.T) {
			if err := validateCommand(s); err == nil {
				t.Errorf("expected error for %q, got nil", s)
			}
		})
	}
}

func TestLoadDirRejectsInjectedCommand(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad.yaml", `
sessions:
  evil:
    windows:
      - panes:
          - command: "rm -rf $HOME"
`)
	_, err := LoadDir(dir)
	if err == nil {
		t.Error("expected error for injected command")
	}
}

func TestLoadDirRejectsWindowCommandInjection(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad.yaml", `
sessions:
  evil:
    windows:
      - name: win1
        panes:
          - command: "ls; rm -rf /"
`)
	_, err := LoadDir(dir)
	if err == nil {
		t.Error("expected error for window command injection")
	}
}

func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
