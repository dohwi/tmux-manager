package setup

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	MarkerStart = "# >>> tmux-manager >>>"
	MarkerEnd   = "# <<< tmux-manager <<<"
)

//go:embed tmux.conf
var defaultTmuxConf []byte

type Options struct {
	BinaryPath string
	HomeDir    string
	ConfigDir  string
	Uninstall  bool
}

var ensureTmuxFn = defaultEnsureTmux

func OverrideEnsureTmux(fn func() error) func() {
	orig := ensureTmuxFn
	ensureTmuxFn = fn
	return func() { ensureTmuxFn = orig }
}

func Setup(opts Options) error {
	home := opts.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home: %w", err)
		}
	}

	binary := opts.BinaryPath
	if binary == "" {
		var err error
		binary, err = os.Executable()
		if err != nil {
			return fmt.Errorf("resolve binary: %w", err)
		}
	}

	cfgDir := opts.ConfigDir
	if cfgDir == "" {
		cfgDir = filepath.Join(home, ".config", "tmux-manager", "sessions")
	}

	if err := linkBinary(home, binary, opts.Uninstall); err != nil {
		return err
	}
	if err := ensureConfigDir(cfgDir, opts.Uninstall); err != nil {
		return err
	}

	if !opts.Uninstall {
		if err := ensureTmuxFn(); err != nil {
			fmt.Fprintf(os.Stderr, "tmux install: %v\n", err)
			fmt.Println("  tmux is required. Install manually: https://github.com/tmux/tmux/wiki")
			return fmt.Errorf("tmux not available and could not be installed automatically")
		}

		cfgParent := filepath.Dir(cfgDir)
		if err := writeTmuxConfigFile(cfgParent); err != nil {
			fmt.Fprintf(os.Stderr, "tmux config file: %v\n", err)
		}

		reloadTmuxFn(cfgParent)
	}

	tmuxConf := filepath.Join(home, ".tmux.conf")
	tmuxConfContent := fmt.Sprintf("source-file %s/tmux.conf", filepath.Dir(cfgDir))
	if err := updateShellFile(tmuxConf, tmuxConfContent, opts.Uninstall); err != nil {
		fmt.Fprintf(os.Stderr, "tmux.conf error: %v\n", err)
	} else if !opts.Uninstall {
		fmt.Println("registered: source-file in ~/.tmux.conf")
	}

	profile := filepath.Join(home, ".profile")
	zshrc := filepath.Join(home, ".zshrc")

	profileContent := "tm restore 2>/dev/null"
	zshrcContent := "alias tm='nocorrect tm'"

	if err := updateShellFile(profile, profileContent, opts.Uninstall); err != nil {
		fmt.Fprintf(os.Stderr, "profile error: %v\n", err)
	} else if !opts.Uninstall {
		fmt.Println("registered: tm restore in ~/.profile")
	}

	if err := updateShellFile(zshrc, zshrcContent, opts.Uninstall); err != nil {
		fmt.Fprintf(os.Stderr, "zshrc error: %v\n", err)
	} else if !opts.Uninstall {
		fmt.Println("registered: nocorrect alias in ~/.zshrc")
	}

	if opts.Uninstall {
		fmt.Println("tmux-manager uninstalled.")
	} else {
		fmt.Println("\ntmux-manager setup complete.")
		fmt.Println("  Run 'tm' to start, or open a new terminal first.")
	}
	return nil
}

func linkBinary(home, binary string, uninstall bool) error {
	binDir := filepath.Join(home, ".local", "bin")
	symlinkPath := filepath.Join(binDir, "tm")

	if uninstall {
		if _, err := os.Lstat(symlinkPath); err == nil {
			if err := os.Remove(symlinkPath); err != nil {
				return fmt.Errorf("remove symlink: %w", err)
			}
			fmt.Printf("removed: %s\n", symlinkPath)
		}
		return nil
	}

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", binDir, err)
	}
	if _, err := os.Lstat(symlinkPath); err == nil {
		fmt.Printf("already linked: %s\n", symlinkPath)
		return nil
	}
	if err := os.Symlink(binary, symlinkPath); err != nil {
		return fmt.Errorf("symlink: %w", err)
	}
	fmt.Printf("linked: %s → %s\n", symlinkPath, binary)
	return nil
}

func ensureConfigDir(dir string, uninstall bool) error {
	if uninstall {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Printf("removed: %s\n", dir)
		} else {
			fmt.Printf("kept: %s (not empty)\n", dir)
		}
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	fmt.Printf("config dir: %s\n", dir)
	return nil
}

var reloadTmuxFn = reloadTmuxConfig

func OverrideReloadTmux(fn func(string)) func() {
	orig := reloadTmuxFn
	reloadTmuxFn = fn
	return func() { reloadTmuxFn = orig }
}

func reloadTmuxConfig(cfgParent string) {
	tmuxConf := filepath.Join(cfgParent, "tmux.conf")
	cmd := exec.Command("tmux", "source-file", tmuxConf)
	if err := cmd.Run(); err != nil {
		return
	}
	fmt.Println("reloaded: tmux config applied")
}

func writeTmuxConfigFile(parentDir string) error {
	path := filepath.Join(parentDir, "tmux.conf")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", parentDir, err)
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("updated: %s\n", path)
	} else {
		fmt.Printf("created: %s\n", path)
	}
	return os.WriteFile(path, defaultTmuxConf, 0o644)
}

func updateShellFile(path, content string, uninstall bool) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existing := string(data)

	if uninstall {
		removed := removeBlock(existing)
		if removed == existing {
			return nil
		}
		return os.WriteFile(path, []byte(removed), 0o644)
	}

	if hasBlock(existing) {
		return nil
	}
	block := fmt.Sprintf("\n%s\n%s\n%s\n", MarkerStart, content, MarkerEnd)
	return os.WriteFile(path, []byte(existing+block), 0o644)
}

func hasBlock(content string) bool {
	return strings.Contains(content, MarkerStart) && strings.Contains(content, MarkerEnd)
}

func removeBlock(content string) string {
	startIdx := strings.Index(content, MarkerStart)
	if startIdx < 0 {
		return content
	}
	endIdx := strings.Index(content[startIdx:], MarkerEnd)
	if endIdx < 0 {
		return content
	}
	endIdx += startIdx + len(MarkerEnd)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}
	return content[:startIdx] + content[endIdx:]
}

func defaultEnsureTmux() error {
	if _, err := exec.LookPath("tmux"); err == nil {
		fmt.Println("tmux: found")
		return nil
	}

	fmt.Print("tmux not found, installing... ")

	pkgManagers := []struct {
		name string
		cmd  []string
	}{
		{"apt", []string{"sudo", "apt-get", "install", "-y", "tmux"}},
		{"brew", []string{"brew", "install", "tmux"}},
		{"dnf", []string{"sudo", "dnf", "install", "-y", "tmux"}},
		{"pacman", []string{"sudo", "pacman", "-S", "--noconfirm", "tmux"}},
		{"zypper", []string{"sudo", "zypper", "install", "-y", "tmux"}},
		{"apk", []string{"sudo", "apk", "add", "tmux"}},
	}

	for _, pm := range pkgManagers {
		if _, err := exec.LookPath(pm.name); err == nil {
			fmt.Printf("via %s\n", pm.name)
			c := exec.Command(pm.cmd[0], pm.cmd[1:]...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("%s install failed: %w", pm.name, err)
			}
			if _, err := exec.LookPath("tmux"); err != nil {
				return fmt.Errorf("tmux still not found after install")
			}
			return nil
		}
	}

	return fmt.Errorf("no supported package manager found (apt/brew/dnf/pacman/zypper/apk)")
}
