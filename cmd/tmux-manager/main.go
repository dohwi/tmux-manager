package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dohwi/tmux-manager/internal/config"
	"github.com/dohwi/tmux-manager/internal/setup"
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
	if err := setup.Setup(setup.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
		os.Exit(1)
	}
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


