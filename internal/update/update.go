package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func FindRepoDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(real)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func CheckUpdate(repoDir string) (bool, error) {
	cmd := exec.Command("git", "-C", repoDir, "fetch", "origin")
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git fetch: %s: %w", strings.TrimSpace(string(out)), err)
	}

	localRaw, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		return false, err
	}
	remoteRaw, err := exec.Command("git", "-C", repoDir, "rev-parse", "origin/main").Output()
	if err != nil {
		return false, err
	}

	local := strings.TrimSpace(string(localRaw))
	remote := strings.TrimSpace(string(remoteRaw))

	return local != remote, nil
}

func Update(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}

	buildCmd := exec.Command("go", "build", "-o", filepath.Join(repoDir, "tmux-manager"), ".")
	buildCmd.Dir = repoDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	return nil
}

func lastCheckPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tmux-manager", ".last-update-check")
}

func ShouldCheck() bool {
	data, err := os.ReadFile(lastCheckPath())
	if err != nil {
		return true
	}
	var ts int64
	fmt.Sscanf(string(data), "%d", &ts)
	return time.Since(time.Unix(ts, 0)) > time.Hour
}

func MarkChecked() {
	os.MkdirAll(filepath.Dir(lastCheckPath()), 0755)
	os.WriteFile(lastCheckPath(), []byte(fmt.Sprintf("%d", time.Now().Unix())), 0644)
}
