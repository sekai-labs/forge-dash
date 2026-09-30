package styles_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestTruncateAndPad(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello world", 5, "he..."},
		{"hello", 10, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"hello", -1, ""},
		{"feature ⚡ fast", 12, "feature ..."},
		{"line1\r\nline2\nline3", 20, "line1 line2 line3"},
		{"   padded   ", 6, "padded"},
	}

	for _, tt := range tests {
		got := styles.Truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q; want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	got := styles.PadRight("test", 8)
	if got != "test    " {
		t.Errorf("PadRight(\"test\", 8) = %q; want %q", got, "test    ")
	}

	gotWide := styles.PadRight("✓", 4)
	if len(gotWide) == 0 {
		t.Errorf("expected non-empty padded string")
	}

	gotExact := styles.PadRight("hello", 5)
	if gotExact != "hello" {
		t.Errorf("PadRight(\"hello\", 5) = %q; want %q", gotExact, "hello")
	}

	gotLess := styles.PadRight("hello world", 5)
	if gotLess != "hello world" {
		t.Errorf("PadRight(\"hello world\", 5) = %q; want %q", gotLess, "hello world")
	}
}

func TestNewTheme_StylesInitialized(t *testing.T) {
	palettes := []struct {
		name string
		p    styles.Palette
	}{
		{"DefaultPalette", styles.DefaultPalette()},
		{"NordPalette", styles.NordPalette()},
	}

	for _, pt := range palettes {
		t.Run(pt.name, func(t *testing.T) {
			theme := styles.NewTheme(pt.p)
			if theme == nil {
				t.Fatalf("expected non-nil theme for %s", pt.name)
			}

			if theme.Title.Render("Test") == "" {
				t.Errorf("Title rendered empty string")
			}
			if theme.Box.Render("Box") == "" {
				t.Errorf("Box rendered empty string")
			}
			if theme.BoxActive.Render("Active") == "" {
				t.Errorf("BoxActive rendered empty string")
			}
			if theme.ActiveBox.Render("Active") == "" {
				t.Errorf("ActiveBox rendered empty string")
			}
			if theme.TableHeader.Render("Header") == "" {
				t.Errorf("TableHeader rendered empty string")
			}
			if theme.TableRow.Render("Row") == "" {
				t.Errorf("TableRow rendered empty string")
			}
			if theme.TableSelected.Render("Selected") == "" {
				t.Errorf("TableSelected rendered empty string")
			}
			if theme.TabActive.Render("Tab") == "" {
				t.Errorf("TabActive rendered empty string")
			}
			if theme.TabInactive.Render("Tab") == "" {
				t.Errorf("TabInactive rendered empty string")
			}
			if theme.ModalBox.Render("Modal") == "" {
				t.Errorf("ModalBox rendered empty string")
			}
			if theme.StatusBar.Render("Status") == "" {
				t.Errorf("StatusBar rendered empty string")
			}
		})
	}
}

func TestTheme_ProviderBadge(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	tests := []struct {
		provider account.ForgeType
		want     string
	}{
		{account.ForgeGitHub, "[GH]"},
		{account.ForgeGitLab, "[GL]"},
		{account.ForgeForgejo, "[CB]"},
		{"unknown", "[--]"},
	}

	for _, tt := range tests {
		got := theme.ProviderBadge(tt.provider)
		if got == "" {
			t.Errorf("ProviderBadge(%v) produced empty output", tt.provider)
		}
	}
}

func TestTheme_ChangeStateIcon(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	states := []change.ChangeState{
		change.ChangeStateOpen,
		change.ChangeStateMerged,
		change.ChangeStateClosed,
		change.ChangeStateDraft,
		"unknown",
	}

	for _, st := range states {
		icon := theme.ChangeStateIcon(st)
		if icon == "" {
			t.Errorf("ChangeStateIcon(%v) returned empty string", st)
		}
	}
}

func TestTheme_IssueStateIcon(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	states := []issue.IssueState{
		issue.IssueStateOpen,
		issue.IssueStateClosed,
		"unknown",
	}

	for _, st := range states {
		icon := theme.IssueStateIcon(st)
		if icon == "" {
			t.Errorf("IssueStateIcon(%v) returned empty string", st)
		}
	}
}
