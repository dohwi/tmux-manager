package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasBlock(t *testing.T) {
	if hasBlock("") {
		t.Error("empty content has no block")
	}
	if hasBlock(MarkerStart) {
		t.Error("only start marker is not a block")
	}
	if !hasBlock(MarkerStart + "\nstuff\n" + MarkerEnd) {
		t.Error("complete block should be detected")
	}
}

func TestRemoveBlock(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no block", "user stuff\n", "user stuff\n"},
		{"removes block", "# stuff\n" + MarkerStart + "\nfoo\n" + MarkerEnd + "\n# after\n", "# stuff\n# after\n"},
		{"preserves prefix", "before\n" + MarkerStart + "\nfoo\n" + MarkerEnd + "\n", "before\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := removeBlock(c.in); got != c.want {
				t.Errorf("removeBlock = %q, want %q", got, c.want)
			}
		})
	}
}

func TestUpdateShellFileIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".profile")

	if err := updateShellFile(path, "tm restore 2>/dev/null", false); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)

	if err := updateShellFile(path, "tm restore 2>/dev/null", false); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("expected idempotent")
	}
}

func TestUpdateShellFileUninstall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")

	if err := updateShellFile(path, "alias tm='nocorrect tm'", false); err != nil {
		t.Fatal(err)
	}
	if err := updateShellFile(path, "alias tm='nocorrect tm'", true); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), MarkerStart) {
		t.Errorf("expected block removed, got %q", got)
	}
}

func TestUpdateShellFilePreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	original := "# user alias\nexport FOO=bar\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateShellFile(path, "alias tm='nocorrect tm'", false); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(got), original) {
		t.Errorf("expected original content preserved at start, got %q", got)
	}
}

func TestSetupFullCycle(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "tmux-manager", "sessions")
	binary := filepath.Join(home, "fake-tmux-manager")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	restore := OverrideEnsureTmux(func() error { return nil })
	defer restore()

	reloadRestore := OverrideReloadTmux(func(string) {})
	defer reloadRestore()

	if err := Setup(Options{BinaryPath: binary, HomeDir: home, ConfigDir: cfgDir}); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(home, ".local", "bin", "tm")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("expected symlink, got %v", err)
	}
	if _, err := os.Stat(cfgDir); err != nil {
		t.Errorf("expected config dir, got %v", err)
	}

	tmuxCfg := filepath.Dir(cfgDir) + "/tmux.conf"
	if _, err := os.Stat(tmuxCfg); err != nil {
		t.Errorf("expected tmux.conf in config dir, got %v", err)
	}

	for _, name := range []string{".profile", ".zshrc", ".tmux.conf"} {
		data, _ := os.ReadFile(filepath.Join(home, name))
		if !hasBlock(string(data)) {
			t.Errorf("expected block in %s, got %q", name, data)
		}
	}

	if err := Setup(Options{BinaryPath: binary, HomeDir: home, ConfigDir: cfgDir}); err != nil {
		t.Fatal(err)
	}

	if err := Setup(Options{BinaryPath: binary, HomeDir: home, ConfigDir: cfgDir, Uninstall: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Errorf("expected symlink removed")
	}
	for _, name := range []string{".profile", ".zshrc", ".tmux.conf"} {
		data, _ := os.ReadFile(filepath.Join(home, name))
		if hasBlock(string(data)) {
			t.Errorf("expected block removed from %s", name)
		}
	}
}

func TestSetupUninstallRefusesToRemoveNonEmptyConfig(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "tmux-manager", "sessions")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "user.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Setup(Options{HomeDir: home, ConfigDir: cfgDir, Uninstall: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfgDir); err != nil {
		t.Errorf("non-empty config dir should be kept")
	}
}

func TestWriteTmuxConfigFile(t *testing.T) {
	dir := t.TempDir()
	if err := writeTmuxConfigFile(dir); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "tmux.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "status-style") {
		t.Errorf("expected status-style in config, got %q", got)
	}

	if err := writeTmuxConfigFile(dir); err != nil {
		t.Fatal(err)
	}
}
