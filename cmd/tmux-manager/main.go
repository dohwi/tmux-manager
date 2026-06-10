package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dohwi/tmux-manager/internal/config"
	"github.com/dohwi/tmux-manager/internal/tmux"
	"github.com/dohwi/tmux-manager/internal/tui"
	"github.com/dohwi/tmux-manager/internal/update"
)

var version string

func main() {
	if len(os.Args) < 2 {
		updateAvailable := checkAutoUpdate()
		runTUI(updateAvailable)
		return
	}

	switch os.Args[1] {
	case "restore":
		runRestore()
	case "setup":
		runSetup()
	case "update":
		runUpdate()
	default:
		updateAvailable := checkAutoUpdate()
		runTUI(updateAvailable)
	}
}

func runRestore() {
	if err := config.RestoreAll(); err != nil {
		fmt.Fprintf(os.Stderr, "restore error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(updateAvailable bool) {
	session, err := tui.Run(updateAvailable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if session == "" {
		return
	}

	if tmux.IsInsideTmux() {
		if err := tmux.SwitchSession(session); err != nil {
			fmt.Fprintf(os.Stderr, "switch error: %v\n", err)
		}
		return
	}

	cmd := exec.Command("tmux", "attach", "-t", session)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "attach error: %v\n", err)
	}
}

func runSetup() {
	binary, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
		os.Exit(1)
	}

	binDir := filepath.Join(home, ".local", "bin")
	symlinkPath := filepath.Join(binDir, "tm")

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir error: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		if err := os.Symlink(binary, symlinkPath); err != nil {
			fmt.Fprintf(os.Stderr, "symlink error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("linked: %s → %s\n", symlinkPath, binary)
	} else {
		fmt.Printf("already linked: %s\n", symlinkPath)
	}

	configDir, _ := config.ConfigDir()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("config dir: %s\n", configDir)

	profile := filepath.Join(home, ".profile")
	if err := ensureLine(profile, "tm restore 2>/dev/null", "# tmux-manager"); err != nil {
		fmt.Fprintf(os.Stderr, "profile error: %v\n", err)
	} else {
		fmt.Println("registered: tm restore in ~/.profile")
	}

	zshrc := filepath.Join(home, ".zshrc")
	_ = ensureLine(zshrc, "alias tm='nocorrect tm'", "# tmux-manager")
	fmt.Println("registered: nocorrect alias in ~/.zshrc")

	fmt.Println("\ntmux-manager setup complete.")
	fmt.Println("  Run 'tm' to start, or open a new terminal first.")
}

func runUpdate() {
	fmt.Fprintln(os.Stderr, "Resolving latest release...")
	_, latest, err := update.CheckUpdate(version, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update check error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Installing %s via go install...\n", latest)
	if err := update.DoUpdate(latest); err != nil {
		fmt.Fprintf(os.Stderr, "update error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Updated successfully. Run 'tm' to start.")
}

func checkAutoUpdate() bool {
	if !update.ShouldCheck() {
		return false
	}

	available, _, err := update.CheckUpdate(version, false)
	update.MarkChecked()
	if err != nil {
		return false
	}

	return available
}

func ensureLine(path, line, marker string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	if strings.Contains(content, line) {
		return nil
	}

	entry := fmt.Sprintf("\n%s\n%s\n", marker, line)
	if err := os.WriteFile(path, append(data, []byte(entry)...), 0o644); err != nil {
		return err
	}
	return nil
}
