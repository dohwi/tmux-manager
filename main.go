package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tmux-manager/internal/config"
	"tmux-manager/internal/tmux"
	"tmux-manager/internal/tui"
	"tmux-manager/internal/update"
)

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

	if err := os.MkdirAll(binDir, 0755); err != nil {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
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
	ensureLine(zshrc, "alias tm='nocorrect tm'", "# tmux-manager")
	fmt.Println("registered: nocorrect alias in ~/.zshrc")

	fmt.Println("\ntmux-manager setup complete.")
	fmt.Println("  Run 'tm' to start, or open a new terminal first.")
}

func runUpdate() {
	repoDir, err := update.FindRepoDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "update error: %v\n", err)
		os.Exit(1)
	}

	needs, err := update.CheckUpdate(repoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check error: %v\n", err)
		os.Exit(1)
	}

	if !needs {
		fmt.Println("Already up to date.")
		update.MarkChecked()
		return
	}

	fmt.Println("Updating...")
	if err := update.Update(repoDir); err != nil {
		fmt.Fprintf(os.Stderr, "update error: %v\n", err)
		os.Exit(1)
	}

	update.MarkChecked()
	fmt.Println("Updated successfully. Run 'tm' to start.")
}

func checkAutoUpdate() bool {
	if !update.ShouldCheck() {
		return false
	}

	repoDir, err := update.FindRepoDir()
	if err != nil {
		return false
	}

	needs, err := update.CheckUpdate(repoDir)
	update.MarkChecked()
	if err != nil {
		return false
	}

	return needs
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
	if err := os.WriteFile(path, append(data, []byte(entry)...), 0644); err != nil {
		return err
	}
	return nil
}
