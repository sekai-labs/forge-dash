package keymap_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/interfaces/tui/keymap"
)

func TestDefaultKeyMap(t *testing.T) {
	km := keymap.DefaultKeyMap()

	if km.Quit.Help().Key != "q" {
		t.Errorf("Quit key help expected 'q', got %q", km.Quit.Help().Key)
	}
	if km.Up.Help().Desc != "up" {
		t.Errorf("Up key help desc expected 'up', got %q", km.Up.Help().Desc)
	}
	if km.Down.Help().Desc != "down" {
		t.Errorf("Down key help desc expected 'down', got %q", km.Down.Help().Desc)
	}
	if km.Top.Help().Desc != "top" {
		t.Errorf("Top key help desc expected 'top', got %q", km.Top.Help().Desc)
	}
	if km.Bottom.Help().Desc != "bottom" {
		t.Errorf("Bottom key help desc expected 'bottom', got %q", km.Bottom.Help().Desc)
	}

	shortHelp := km.ShortHelp()
	if len(shortHelp) == 0 {
		t.Fatalf("expected non-empty ShortHelp")
	}

	groups := km.HelpGroups()
	if len(groups) == 0 {
		t.Fatalf("expected non-empty HelpGroups")
	}

	expectedGroupTitles := []string{
		"Navigation",
		"Views & Sections",
		"Panels & Focus",
		"Preview Pane",
		"Actions",
		"Modals & Global",
	}

	for i, expected := range expectedGroupTitles {
		if i >= len(groups) {
			t.Errorf("missing group %q at index %d", expected, i)
			continue
		}
		if groups[i].Title != expected {
			t.Errorf("group %d title: got %q, want %q", i, groups[i].Title, expected)
		}
		if len(groups[i].Bindings) == 0 {
			t.Errorf("group %q has no bindings", groups[i].Title)
		}
	}
}
