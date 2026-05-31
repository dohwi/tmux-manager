package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"tmux-manager/internal/tmux"
)

type PaneConfig struct {
	Command   string `yaml:"command,omitempty"`
	Direction string `yaml:"direction,omitempty"` // "right" | "down"
}

type WindowConfig struct {
	Name    string       `yaml:"name,omitempty"`
	Command string       `yaml:"command,omitempty"`
	Panes   []PaneConfig `yaml:"panes,omitempty"`
}

type SessionConfig struct {
	Name      string         `yaml:"name"`
	Directory string         `yaml:"directory,omitempty"`
	Command   string         `yaml:"command,omitempty"`
	Windows   []WindowConfig `yaml:"windows,omitempty"`
	Panes     []PaneConfig   `yaml:"panes,omitempty"`
}

type configFile struct {
	Sessions []SessionConfig `yaml:"sessions"`
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

func RestoreAll() error {
	configs, err := LoadAll()
	if err != nil {
		return err
	}

	for name, cfg := range configs {
		if tmux.HasSession(name) {
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
	if cfg.Directory != "" {
		args = append(args, "-c", expandHome(cfg.Directory))
	}

	if len(cfg.Windows) > 0 {
		args = append(args, "-n", cfg.Windows[0].Name)
	}

	if err := exec.Command("tmux", args...).Run(); err != nil {
		return err
	}

	if len(cfg.Windows) > 0 {
		createWindowPanes(name, 0, cfg.Windows[0])
	} else if len(cfg.Panes) > 0 {
		createWindowPanes(name, 0, WindowConfig{Panes: cfg.Panes})
	} else if cfg.Command != "" {
		exec.Command("tmux", "send-keys", "-t", name+":0", cfg.Command, "Enter").Run()
	}

	for i := 1; i < len(cfg.Windows); i++ {
		w := cfg.Windows[i]
		exec.Command("tmux", "new-window", "-t", name, "-n", w.Name).Run()
		createWindowPanes(name, i, w)
	}

	return nil
}

func createWindowPanes(session string, windowIdx int, w WindowConfig) {
	target := fmt.Sprintf("%s:%d", session, windowIdx)

	if len(w.Panes) > 0 {
		if w.Panes[0].Command != "" {
			exec.Command("tmux", "send-keys", "-t", target+".0", w.Panes[0].Command, "Enter").Run()
		}
		for i := 1; i < len(w.Panes); i++ {
			flag := "-v"
			if w.Panes[i].Direction == "right" {
				flag = "-h"
			}
			exec.Command("tmux", "split-window", "-t", target, flag).Run()
			if w.Panes[i].Command != "" {
				exec.Command("tmux", "send-keys", "-t", target, w.Panes[i].Command, "Enter").Run()
			}
		}
		return
	}

	if w.Command != "" {
		exec.Command("tmux", "send-keys", "-t", target+".0", w.Command, "Enter").Run()
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
