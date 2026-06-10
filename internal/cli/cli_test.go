package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	<-done
	return buf.String(), runErr
}

func TestRootCmdHelp(t *testing.T) {
	root := newRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
}

func TestVersionCmd(t *testing.T) {
	root := newRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"version"})

	out, err := captureStdout(t, func() error {
		r := newRootCmd()
		r.SetArgs([]string{"version"})
		r.SetOut(os.Stdout)
		r.SetErr(io.Discard)
		return r.Execute()
	})
	if err != nil {
		t.Fatalf("version cmd failed: %v", err)
	}
	if !strings.Contains(out, "tmux-manager") {
		t.Errorf("expected version output, got %q", out)
	}
}

func TestRestoreDryRunNoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := newRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"restore", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("restore --dry-run failed: %v", err)
	}
}

func TestConfirmAcceptsY(t *testing.T) {
	old := os.Stdin
	defer func() { os.Stdin = old }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.Write([]byte("y\n"))
		_ = w.Close()
	}()
	if !confirm() {
		t.Error("expected confirm=true for 'y'")
	}
}

func TestConfirmRejectsEmpty(t *testing.T) {
	old := os.Stdin
	defer func() { os.Stdin = old }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_ = w.Close()
	}()
	if confirm() {
		t.Error("expected confirm=false for empty input")
	}
}
