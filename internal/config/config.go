package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/dohwi/tmux-manager/internal/tmux"
)

type PaneConfig struct {
	Command   string `yaml:"command,omitempty"`
	Direction string `yaml:"direction,omitempty"` // "right" | "down"
	Directory string `yaml:"directory,omitempty"`
	Name      string `yaml:"name,omitempty"`
}

type WindowConfig struct {
	Name  string       `yaml:"name,omitempty"`
	Panes []PaneConfig `yaml:"panes,omitempty"`
}

type SessionConfig struct {
	Windows []WindowConfig `yaml:"windows,omitempty"`
}

type configFile struct {
	Sessions map[string]SessionConfig `yaml:"sessions"`
}

// execCommand is overridable for tests; same shape as exec.Command.
var execCommand = exec.Command

// hasSessionFn is overridable for tests.
var hasSessionFn = tmux.HasSession

func validateSessionCommands(_ string, session SessionConfig) error {
	for _, w := range session.Windows {
		for _, p := range w.Panes {
			if err := validateCommand(p.Command); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadDir reads YAML files from the given directory and returns parsed configs.
func LoadDir(dir string) (map[string]SessionConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	configs := make(map[string]SessionConfig)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var file configFile
		if err := yaml.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		for name, cfg := range file.Sessions {
			if name == "" {
				return nil, fmt.Errorf("%s: session name must not be empty", entry.Name())
			}
			if _, ok := configs[name]; ok {
				return nil, fmt.Errorf("duplicate session name: %s (from %s)", name, entry.Name())
			}
			if err := validateSessionCommands(name, cfg); err != nil {
				return nil, fmt.Errorf("%s session %q: %w", entry.Name(), name, err)
			}
			configs[name] = cfg
		}
	}
	return configs, nil
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tmux-manager", "sessions"), nil
}

func LoadAll() (map[string]SessionConfig, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	return LoadDir(dir)
}

func RestoreAll() error {
	configs, err := LoadAll()
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		dir, _ := ConfigDir()
		fmt.Fprintf(os.Stderr, "no sessions defined. Add YAML files to %s\n", dir)
		return nil
	}

	for name, cfg := range configs {
		if hasSessionFn(name) {
			continue
		}
		if err := createSession(name, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create session %s: %v\n", name, err)
			continue
		}
		fmt.Printf("created session: %s\n", name)
	}
	return nil
}

func createSession(name string, cfg SessionConfig) error {
	args := []string{"new-session", "-d", "-s", name, "-P", "-F", "#{window_index}.#{pane_index}"}

	if len(cfg.Windows) > 0 {
		if cfg.Windows[0].Name != "" {
			args = append(args, "-n", cfg.Windows[0].Name)
		} else {
			args = append(args, "-n", name)
		}
	} else {
		args = append(args, "-n", name)
	}

	out, err := execCommand("tmux", args...).Output()
	if err != nil {
		return err
	}
	firstWinIdx, firstPaneIdx := parseWindowPaneIndex(strings.TrimSpace(string(out)))

	if len(cfg.Windows) > 0 {
		if err := createWindowPanes(name, firstWinIdx, firstPaneIdx, cfg.Windows[0]); err != nil {
			return err
		}
	}

	for i := 1; i < len(cfg.Windows); i++ {
		w := cfg.Windows[i]
		out, err := execCommand("tmux", "new-window", "-t", name, "-n", w.Name, "-P", "-F", "#{window_index}.#{pane_index}").Output()
		if err != nil {
			return fmt.Errorf("new-window %s:%d: %w", name, i, err)
		}
		winIdx, paneIdx := parseWindowPaneIndex(strings.TrimSpace(string(out)))
		if err := createWindowPanes(name, winIdx, paneIdx, w); err != nil {
			return err
		}
	}

	return nil
}

func createWindowPanes(session string, windowIdx, firstPaneIdx int, w WindowConfig) error {
	if len(w.Panes) == 0 {
		return nil
	}
	target := fmt.Sprintf("%s:%d", session, windowIdx)
	paneIdx := firstPaneIdx
	for i, pane := range w.Panes {
		if i > 0 {
			flag := "-v"
			if pane.Direction == "right" {
				flag = "-h"
			}
			out, err := execCommand("tmux", "split-window", "-t", target, flag, "-P", "-F", "#{pane_index}").Output()
			if err != nil {
				return fmt.Errorf("split-window %s.%d: %w", target, i, err)
			}
			paneIdx, _ = strconv.Atoi(strings.TrimSpace(string(out)))
		}
		paneTarget := fmt.Sprintf("%s.%d", target, paneIdx)
		if pane.Name != "" {
			if err := execCommand("tmux", "select-pane", "-t", paneTarget, "-T", pane.Name).Run(); err != nil {
				return fmt.Errorf("select-pane %s: %w", paneTarget, err)
			}
		}
		cmd := paneCommand(pane)
		if cmd != "" {
			if err := execCommand("tmux", "send-keys", "-t", paneTarget, cmd, "Enter").Run(); err != nil {
				return fmt.Errorf("send-keys %s: %w", paneTarget, err)
			}
		}
	}
	return nil
}

// parseWindowPaneIndex parses a "window_index.pane_index" string like "1.1".
func parseWindowPaneIndex(s string) (winIdx, paneIdx int) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) >= 1 {
		winIdx, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		paneIdx, _ = strconv.Atoi(parts[1])
	}
	return
}

func paneCommand(p PaneConfig) string {
	cmd := p.Command
	if p.Directory != "" {
		cd := fmt.Sprintf("cd %s", expandHome(p.Directory))
		if cmd != "" {
			cmd = fmt.Sprintf("%s && %s", cd, cmd)
		} else {
			cmd = cd
		}
	}
	return cmd
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
