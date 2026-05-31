package main

import (
	"fmt"
	"os"
	"os/exec"

	"tmux-manager/internal/config"
	"tmux-manager/internal/tmux"
	"tmux-manager/internal/tui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "restore" {
		if err := config.RestoreAll(); err != nil {
			fmt.Fprintf(os.Stderr, "restore error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	session, err := tui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if session == "" {
		return
	}

	if err := tmux.SwitchSession(session); err == nil {
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
