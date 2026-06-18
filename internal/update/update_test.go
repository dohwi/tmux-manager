package update

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeTag(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"v1.2.3", "1.2.3"},
		{"  v0.0.1  ", "0.0.1"},
		{"1.2.3", "1.2.3"},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalizeTag(c.in); got != c.want {
			t.Errorf("normalizeTag(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.2.3", "1.10.0", -1},
		{"2.0.0", "1.99.99", 1},
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-rc.1", "1.0.0", -1},
		{"1.0.0-rc.1", "1.0.0-rc.2", -1},
		{"1.0.0-rc.2", "1.0.0-rc.1", 1},
		{"1.0.0-rc.1", "1.0.0-rc.1", 0},
		{"1.0.0-alpha", "1.0.0-1", 1},
		{"1.0.0-1", "1.0.0-alpha", -1},
		{"1.0.0+build1", "1.0.0+build2", 0},
		{"1.0.0+build1", "1.0.1", -1},
	}
	for _, c := range cases {
		got, err := compareSemver(c.a, c.b)
		if err != nil {
			t.Errorf("compareSemver(%q, %q) error: %v", c.a, c.b, err)
			continue
		}
		if got != c.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCheckUpdateEmptyVersion(t *testing.T) {
	available, tag, err := CheckUpdate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available || tag != "" {
		t.Errorf("expected (false, \"\"), got (%v, %q)", available, tag)
	}
}

func TestCheckUpdateNewerAvailable(t *testing.T) {
	restore := swapFetcher(func() (string, error) { return "v0.99.0", nil })
	defer restore()

	available, tag, err := CheckUpdate("v0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected update available")
	}
	if tag != "v0.99.0" {
		t.Errorf("expected tag=v0.99.0, got %q", tag)
	}
}

func TestCheckUpdateSameVersion(t *testing.T) {
	restore := swapFetcher(func() (string, error) { return "v0.1.0", nil })
	defer restore()

	available, tag, err := CheckUpdate("v0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected no update")
	}
	if tag != "v0.1.0" {
		t.Errorf("expected tag=v0.1.0, got %q", tag)
	}
}

func TestCheckUpdateOlderLocal(t *testing.T) {
	restore := swapFetcher(func() (string, error) { return "v0.1.0", nil })
	defer restore()

	available, _, err := CheckUpdate("v0.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected no update (local is newer)")
	}
}

func TestCheckUpdateFetcherError(t *testing.T) {
	restore := swapFetcher(func() (string, error) { return "", errors.New("network down") })
	defer restore()

	_, _, err := CheckUpdate("v0.1.0")
	if err == nil {
		t.Error("expected error from fetcher")
	}
}

func TestDefaultFetcherNonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	oldURL := githubAPIURL
	githubAPIURL = srv.URL
	defer func() { githubAPIURL = oldURL }()

	_, err := defaultFetcher()
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestShouldCheckAndMarkChecked(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if !ShouldCheck() {
		t.Error("expected ShouldCheck=true on empty state")
	}

	MarkChecked()
	if ShouldCheck() {
		t.Error("expected ShouldCheck=false after MarkChecked")
	}

	old := fmt.Sprintf("%d", time.Now().Add(-2*time.Hour).Unix())
	path := filepath.Join(dir, ".config", "tmux-manager", ".last-update-check")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if !ShouldCheck() {
		t.Error("expected ShouldCheck=true for stale timestamp")
	}
}

func TestDoUpdateBuildsSpec(t *testing.T) {
	origExec := execCommand
	execCommand = fakeExec(t, "go", "install", cmdPath+"@v1.2.3")
	defer func() { execCommand = origExec }()

	if err := DoUpdate("v1.2.3"); err == nil {
		t.Error("expected error from fake exec, got nil")
	}
}

func TestDoUpdateSetsGOBINToBinaryDir(t *testing.T) {
	dir := t.TempDir()
	fakeBin := filepath.Join(dir, "tmux-manager")
	if err := os.WriteFile(fakeBin, []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	origExec := execCommand
	defer func() { execCommand = origExec }()
	origExecutable := executableFn
	defer func() { executableFn = origExecutable }()

	executableFn = func() (string, error) { return fakeBin, nil }

	envFile := filepath.Join(dir, "env.txt")
	script := filepath.Join(dir, "fake-go")
	scriptContent := "#!/bin/sh\nenv > " + envFile + "\nexit 1\n"
	if err := os.WriteFile(script, []byte(scriptContent), 0o755); err != nil {
		t.Fatal(err)
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name != "go" {
			t.Errorf("exec name = %q, want go", name)
		}
		return exec.Command(script)
	}

	if err := DoUpdate("v1.2.3"); err == nil {
		t.Error("expected error from fake exec, got nil")
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("fake go did not write env file: %v", err)
	}

	want := "GOBIN=" + dir
	if !strings.Contains(string(data), want) {
		t.Errorf("expected env to contain %q, got:\n%s", want, data)
	}
}

func swapFetcher(fn func() (string, error)) func() {
	orig := fetcher
	fetcher = fn
	return func() { fetcher = orig }
}
