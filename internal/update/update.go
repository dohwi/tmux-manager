package update

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	modulePath = "github.com/dohwi/tmux-manager"
	cmdPath    = modulePath + "/cmd/tmux-manager"
)

func CheckUpdate(currentVersion string) (bool, error) {
	if currentVersion == "" {
		return false, nil
	}

	cmd := exec.Command("go", "list", "-m", "-json", modulePath+"@main")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("go list: %w", err)
	}

	var info struct {
		Origin struct {
			Hash string
		}
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return false, fmt.Errorf("parse go list output: %w", err)
	}

	return info.Origin.Hash != currentVersion, nil
}

func DoUpdate() error {
	cmd := exec.Command("go", "install", cmdPath+"@main")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go install: %s", strings.TrimSpace(string(out)))
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
	_, _ = fmt.Sscanf(string(data), "%d", &ts)
	return time.Since(time.Unix(ts, 0)) > time.Hour
}

func MarkChecked() {
	_ = os.MkdirAll(filepath.Dir(lastCheckPath()), 0o755)
	_ = os.WriteFile(lastCheckPath(), []byte(fmt.Sprintf("%d", time.Now().Unix())), 0o644)
}
