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

type WindowConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command,omitempty"`
}

type SessionConfig struct {
	Directory string         `yaml:"directory,omitempty"`
	Command   string         `yaml:"command,omitempty"`
	Windows   []WindowConfig `yaml:"windows,omitempty"`
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
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		fullPath := filepath.Join(dir, entry.Name())

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var cfg SessionConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		configs[name] = cfg
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

	sessionCmd := cfg.Command
	firstWindowCmd := ""
	if len(cfg.Windows) > 0 {
		firstWindowCmd = cfg.Windows[0].Command
	}

	if firstWindowCmd != "" {
		exec.Command("tmux", "send-keys", "-t", name+":0", firstWindowCmd, "Enter").Run()
	} else if sessionCmd != "" {
		exec.Command("tmux", "send-keys", "-t", name+":0", sessionCmd, "Enter").Run()
	}

	for i := 1; i < len(cfg.Windows); i++ {
		w := cfg.Windows[i]
		exec.Command("tmux", "new-window", "-t", name, "-n", w.Name).Run()
		if w.Command != "" {
			exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:%d", name, i), w.Command, "Enter").Run()
		}
	}

	return nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
