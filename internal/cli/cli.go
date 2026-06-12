package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dohwi/tmux-manager/internal/config"
	"github.com/dohwi/tmux-manager/internal/setup"
	"github.com/dohwi/tmux-manager/internal/tmux"
	"github.com/dohwi/tmux-manager/internal/tui"
	"github.com/dohwi/tmux-manager/internal/update"
)

var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0") {
			Version = v
			return
		}
	}
	if v := resolveGitVersion(); v != "" {
		Version = v
	}
}

func resolveGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--dirty", "--always")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tmux-manager",
		Short:         "TUI session manager for tmux with YAML-defined workspaces",
		Long:          "tmux-manager launches a TUI to browse, attach, and manage tmux sessions.\nRun with no arguments for the TUI, or use a subcommand.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
		RunE:          runTUI,
	}
	root.SetVersionTemplate("tmux-manager {{.Version}}\n")

	root.AddCommand(
		newRestoreCmd(),
		newSetupCmd(),
		newUpdateCmd(),
		newVersionCmd(),
	)
	return root
}

func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func runTUI(cmd *cobra.Command, _ []string) error {
	updateAvailable := checkAutoUpdate()
	return launchTUI(updateAvailable)
}

func checkAutoUpdate() bool {
	if isDev() {
		return false
	}
	if !update.ShouldCheck() {
		return update.IsCachedUpdateAvailable()
	}
	available, _, err := update.CheckUpdate(Version)
	update.MarkChecked()
	if err != nil {
		return update.IsCachedUpdateAvailable()
	}
	update.CacheUpdateState(available)
	return available
}

func launchTUI(updateAvailable bool) error {
	session, err := tui.Run(updateAvailable)
	if err != nil {
		return err
	}
	if session == "" {
		return nil
	}
	if tmux.IsInsideTmux() {
		return tmux.SwitchSession(session)
	}
	c := exec.Command("tmux", "attach", "-t", session)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("attach: %w", err)
	}
	return nil
}

func newRestoreCmd() *cobra.Command {
	var dryRun bool
	c := &cobra.Command{
		Use:   "restore",
		Short: "Restore sessions from YAML workspace definitions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dryRun {
				configs, err := config.LoadAll()
				if err != nil {
					return err
				}
				fmt.Printf("Would restore %d session(s):\n", len(configs))
				for name := range configs {
					fmt.Printf("  - %s\n", name)
				}
				return nil
			}
			return config.RestoreAll()
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be restored without creating sessions")
	return c
}

func newSetupCmd() *cobra.Command {
	var uninstall bool
	c := &cobra.Command{
		Use:   "setup",
		Short: "Install symlink and shell integration (or remove with --uninstall)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return setup.Setup(setup.Options{Uninstall: uninstall})
		},
	}
	c.Flags().BoolVar(&uninstall, "uninstall", false, "remove symlink and shell integration")
	return c
}

func newUpdateCmd() *cobra.Command {
	var (
		checkOnly bool
		yes       bool
	)
	c := &cobra.Command{
		Use:   "update",
		Short: "Update tmux-manager to the latest release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if isDev() {
				fmt.Println("Development build. Use 'go install github.com/dohwi/tmux-manager/cmd/tmux-manager@latest' to update.")
				return nil
			}
			available, latest, err := update.CheckUpdate(Version)
			if err != nil {
				return fmt.Errorf("check: %w", err)
			}
			if !available {
				fmt.Printf("Already up to date (%s)\n", latest)
				return nil
			}
			fmt.Printf("Update available: %s -> %s\n", Version, latest)
			if checkOnly {
				return nil
			}
			if !yes && !confirm() {
				fmt.Println("Cancelled.")
				return nil
			}
			fmt.Printf("Installing %s...\n", latest)
			if err := update.DoUpdate(latest); err != nil {
				return err
			}
			fmt.Println("Updated successfully. Run 'tm' to start.")
			update.ClearUpdateCache()
			return nil
		},
	}
	c.Flags().BoolVar(&checkOnly, "check", false, "only check for updates, do not install")
	c.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return c
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tmux-manager version",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Printf("tmux-manager %s\n", Version)
		},
	}
}

func confirm() bool {
	fmt.Printf("Proceed? [y/N] ")
	var ans string
	_, _ = fmt.Scanln(&ans)
	return ans == "y" || ans == "Y"
}

func isDev() bool {
	switch {
	case Version == "dev", Version == "(devel)":
		return true
	case strings.HasPrefix(Version, "v0.0.0"):
		return true
	case strings.Contains(Version, "-dirty"):
		return true
	case strings.Contains(Version, "-g"):
		return true
	case !strings.HasPrefix(Version, "v"):
		return true
	}
	return false
}
