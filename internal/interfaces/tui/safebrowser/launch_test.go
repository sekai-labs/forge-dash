package safebrowser

import (
	"path/filepath"
	"testing"
)

func TestWindowsBrowserLaunchPassesURLAsOneArgument(t *testing.T) {
	const rawURL = "https://forge.example/item?x=1&whoami"
	u, err := ValidateURL(rawURL)
	if err != nil {
		t.Fatalf("expected valid query URL, got %v", err)
	}

	cmd := browserCommand("windows", u.String())
	if filepath.Base(cmd.Path) != "rundll32.exe" {
		t.Fatalf("Windows launcher = %q, want rundll32.exe", cmd.Path)
	}
	if len(cmd.Args) != 3 || cmd.Args[2] != rawURL {
		t.Fatalf("URL was not passed as one argument: %#v", cmd.Args)
	}
}
