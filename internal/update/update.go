package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	githubRepo     = "dohwi/tmux-manager"
	githubAPIURL   = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	throttleWindow = time.Hour
)

var (
	modulePath = "github.com/" + githubRepo
	cmdPath    = modulePath + "/cmd/tmux-manager"
	nowFunc    = time.Now
)

var execCommand = exec.Command

// fetcher abstracts the network call so tests can inject a response.
var fetcher = defaultFetcher

var httpClient = &http.Client{Timeout: 5 * time.Second}

func defaultFetcher() (string, error) {
	req, err := http.NewRequest(http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "tmux-manager")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload struct {
		TagName    string `json:"tag_name"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("parse github response: %w", err)
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}
	if payload.Prerelease {
		return "", fmt.Errorf("latest release is prerelease")
	}
	return payload.TagName, nil
}

func CheckUpdate(currentVersion string) (bool, string, error) {
	current := normalizeTag(currentVersion)
	if current == "" {
		return false, "", nil
	}

	tag := current
	if fetcher != nil {
		latest, err := fetcher()
		if err != nil {
			return false, "", err
		}
		tag = normalizeTag(latest)
	}

	if tag == "" {
		return false, "", nil
	}

	cmp, err := compareSemver(current, tag)
	if err != nil {
		return false, tag, err
	}
	return cmp < 0, tag, nil
}

func DoUpdate(tag string) error {
	spec := cmdPath + "@latest"
	if tag != "" {
		spec = cmdPath + "@" + strings.TrimPrefix(tag, "v")
	}

	cmd := execCommand("go", "install", spec)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go install %s: %s", spec, strings.TrimSpace(string(out)))
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
	ts, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return true
	}
	return nowFunc().Sub(time.Unix(ts, 0)) > throttleWindow
}

func MarkChecked() {
	dir := filepath.Dir(lastCheckPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(lastCheckPath(), []byte(strconv.FormatInt(nowFunc().Unix(), 10)), 0o644)
}
