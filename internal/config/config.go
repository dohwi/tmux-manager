package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	Name      string       `yaml:"name,omitempty"`
	Command   string       `yaml:"command,omitempty"`
	Directory string       `yaml:"directory,omitempty"`
	Panes     []PaneConfig `yaml:"panes,omitempty"`
}

type SessionConfig struct {
	Name    string         `yaml:"name"`
	Windows []WindowConfig `yaml:"windows,omitempty"`
	Panes   []PaneConfig   `yaml:"panes,omitempty"`
}

type configFile struct {
	Sessions []SessionConfig `yaml:"sessions"`
}

// execCommand is overridable for tests; same shape as exec.Command.
var execCommand = exec.Command

// hasSessionFn is overridable for tests.
var hasSessionFn = tmux.HasSession

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

		for _, cfg := range file.Sessions {
			if cfg.Name == "" {
				return nil, fmt.Errorf("%s: each session must have a 'name' field", entry.Name())
			}
			if _, ok := configs[cfg.Name]; ok {
				return nil, fmt.Errorf("duplicate session name: %s (from %s)", cfg.Name, entry.Name())
			}
			configs[cfg.Name] = cfg
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
	args := []string{"new-session", "-d", "-s", name}

	if len(cfg.Windows) > 0 {
		args = append(args, "-n", cfg.Windows[0].Name)
	}

	if err := execCommand("tmux", args...).Run(); err != nil {
		return err
	}

	if len(cfg.Windows) > 0 {
		createWindowPanes(name, 0, cfg.Windows[0])
	} else if len(cfg.Panes) > 0 {
		createWindowPanes(name, 0, WindowConfig{Panes: cfg.Panes})
	}

	for i := 1; i < len(cfg.Windows); i++ {
		w := cfg.Windows[i]
		_ = execCommand("tmux", "new-window", "-t", name, "-n", w.Name).Run()
		createWindowPanes(name, i, w)
	}

	return nil
}

func createWindowPanes(session string, windowIdx int, w WindowConfig) {
	target := fmt.Sprintf("%s:%d", session, windowIdx)

	if len(w.Panes) > 0 {
		for i, pane := range w.Panes {
			if i > 0 {
				flag := "-v"
				if pane.Direction == "right" {
					flag = "-h"
				}
				_ = execCommand("tmux", "split-window", "-t", target, flag).Run()
			}
			paneTarget := fmt.Sprintf("%s.%d", target, i)
			if pane.Name != "" {
				_ = execCommand("tmux", "select-pane", "-t", paneTarget, "-T", pane.Name).Run()
			}
			cmd := paneCommand(pane)
			if cmd != "" {
				_ = execCommand("tmux", "send-keys", "-t", paneTarget, cmd, "Enter").Run()
			}
		}
		return
	}

	if w.Command != "" {
		cmd := w.Command
		if w.Directory != "" {
			cmd = fmt.Sprintf("cd %s && %s", expandHome(w.Directory), cmd)
		}
		_ = execCommand("tmux", "send-keys", "-t", target+".0", cmd, "Enter").Run()
	}
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
