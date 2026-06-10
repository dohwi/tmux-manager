package update

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCheckUpdateSameVersion(t *testing.T) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	hash := strings.TrimSpace(string(out))

	available, err := CheckUpdate(hash)
	if err != nil {
		t.Fatal(err)
	}
	if available {
		t.Error("expected no update (same version)")
	}
}

func TestCheckUpdateEmptyVersion(t *testing.T) {
	available, err := CheckUpdate("")
	if err != nil {
		t.Fatal(err)
	}
	if available {
		t.Error("expected no update for empty version")
	}
}

func TestCheckUpdateDifferentVersion(t *testing.T) {
	available, err := CheckUpdate("0000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if !available {
		t.Error("expected update available for different version")
	}
}

func TestDoUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go install in short mode")
	}
	err := DoUpdate()
	if err != nil {
		t.Fatal(err)
	}
}
