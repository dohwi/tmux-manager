package update

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCheckUpdateEmptyVersion(t *testing.T) {
	available, tag, err := CheckUpdate("", false)
	if err != nil {
		t.Fatal(err)
	}
	if available || tag != "" {
		t.Errorf("expected (false, \"\"), got (%v, %q)", available, tag)
	}
}

func TestCheckUpdateSameVersion(t *testing.T) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		t.Skip("not in a git repo")
	}
	hash := strings.TrimSpace(string(out))
	if hash == "" {
		t.Skip("no git hash available")
	}

	available, _, err := CheckUpdate(hash, false)
	if err != nil {
		t.Skipf("network unavailable: %v", err)
	}
	if available {
		t.Error("expected no update (same version)")
	}
}
