package modals_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components/modals"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func modalLines(t *testing.T, rendered string) []string {
	t.Helper()
	stripped := ansi.Strip(rendered)
	stripped = strings.TrimSuffix(stripped, "\n")
	if stripped == "" {
		t.Fatalf("expected modal output")
	}
	return strings.Split(stripped, "\n")
}

func assertModalBounds(t *testing.T, rendered string, maxWidth, maxHeight int) []string {
	t.Helper()
	lines := modalLines(t, rendered)
	if len(lines) > maxHeight {
		t.Fatalf("modal rendered %d lines, exceeds height %d:\n%s", len(lines), maxHeight, strings.Join(lines, "\n"))
	}
	for i, line := range lines {
		if got := ansi.StringWidth(line); got > maxWidth {
			t.Fatalf("modal line %d is %d cells wide, exceeds %d: %q", i, got, maxWidth, line)
		}
	}
	return lines
}

func testAccounts(n int) []account.Account {
	accounts := make([]account.Account, 0, n)
	for i := 1; i <= n; i++ {
		accounts = append(accounts, account.Account{
			ID:       account.AccountID(fmt.Sprintf("acc-%02d", i)),
			Name:     fmt.Sprintf("Account %02d", i),
			Provider: account.ForgeGitHub,
			Host:     "github.com",
		})
	}
	return accounts
}

func TestModals_HelpStaysWithinViewportAndPagesGroups(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	expectedPages := [][]string{
		{"NAVIGATION", "VIEWS & SECTIONS"},
		{"PANELS & FOCUS", "PREVIEW PANE"},
		{"ACTIONS", "MODALS & GLOBAL"},
	}

	seen := map[string]bool{}
	for page, wantGroups := range expectedPages {
		out := modals.RenderHelpModal(theme, 80, 24, page)
		lines := assertModalBounds(t, out, 76, 22)
		joined := strings.Join(lines, "\n")
		for _, group := range wantGroups {
			if !strings.Contains(joined, group) {
				t.Fatalf("help page %d missing group %q:\n%s", page+1, group, joined)
			}
			seen[group] = true
		}
		if !strings.Contains(joined, fmt.Sprintf("Page %d/3", page+1)) {
			t.Fatalf("help page %d missing footer:\n%s", page+1, joined)
		}
		if !strings.Contains(joined, "←/→") {
			t.Fatalf("help page %d missing page navigation hint:\n%s", page+1, joined)
		}
	}

	if len(seen) != 6 {
		t.Fatalf("expected all six help groups across pages, saw %d", len(seen))
	}

	wide := modals.RenderHelpModal(theme, 120, 40, 0)
	wideLines := assertModalBounds(t, wide, 116, 38)
	wideJoined := strings.Join(wideLines, "\n")
	for _, group := range []string{
		"NAVIGATION",
		"VIEWS & SECTIONS",
		"PANELS & FOCUS",
		"PREVIEW PANE",
		"ACTIONS",
		"MODALS & GLOBAL",
	} {
		if !strings.Contains(wideJoined, group) {
			t.Fatalf("wide help modal missing group %q:\n%s", group, wideJoined)
		}
	}
	if strings.Contains(wideJoined, "Page 1/3") {
		t.Fatalf("wide help modal should not be paged:\n%s", wideJoined)
	}
}

func TestModals_AccountListWindowsAroundSelection(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	accounts := testAccounts(20)

	out := modals.RenderAccountModal(theme, accounts, "acc-05", false, 10, 80, 24)
	lines := assertModalBounds(t, out, 76, 22)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Account 10") {
		t.Fatalf("selected account Account 10 was not windowed into view:\n%s", joined)
	}
	if strings.Contains(joined, "Unified View (All Accounts)") {
		t.Fatalf("long account list should window away from the unified option:\n%s", joined)
	}
	var selected string
	for _, line := range lines {
		if strings.Contains(line, ">") {
			selected = line
		}
	}
	if !strings.Contains(selected, "Account 10") {
		t.Fatalf("expected Account 10 to be the selected row, got %q", selected)
	}

	out = modals.RenderAccountModal(theme, accounts, "acc-05", false, 20, 80, 24)
	lines = assertModalBounds(t, out, 76, 22)
	joined = strings.Join(lines, "\n")
	if !strings.Contains(joined, "Account 20") {
		t.Fatalf("last account was not windowed into view:\n%s", joined)
	}
	var lastSelected string
	for _, line := range lines {
		if strings.Contains(line, ">") {
			lastSelected = line
		}
	}
	if !strings.Contains(lastSelected, "Account 20") {
		t.Fatalf("expected Account 20 to be selected, got %q", lastSelected)
	}

	out = modals.RenderAccountModal(theme, accounts, "acc-05", false, 18, 40, 12)
	lines = assertModalBounds(t, out, 36, 10)
	joined = strings.Join(lines, "\n")
	if !strings.Contains(joined, "Account 18") {
		t.Fatalf("small account modal did not window the selection:\n%s", joined)
	}
}

func TestModals_ConfirmWrapsLongPromptAndKeepsFooter(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	prompt := "Merge PR #142: " + strings.Repeat("this change touches the renderer carefully ", 6) + "?"

	out := modals.RenderConfirmModal(theme, "Confirmation Required", prompt, 80, 24)
	lines := assertModalBounds(t, out, 76, 22)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "#142") {
		t.Fatalf("confirm modal lost the item ID:\n%s", joined)
	}
	if !strings.Contains(joined, "[y]") {
		t.Fatalf("confirm modal lost the action footer:\n%s", joined)
	}
	if !strings.Contains(joined, "Confirmation Required") {
		t.Fatalf("confirm modal lost its title:\n%s", joined)
	}

	small := modals.RenderConfirmModal(theme, "Merge PR #142", "Merge PR #142: "+strings.Repeat("long prompt ", 2), 40, 16)
	smallLines := assertModalBounds(t, small, 36, 14)
	smallJoined := strings.Join(smallLines, "\n")
	if !strings.Contains(smallJoined, "#142") {
		t.Fatalf("small confirm modal lost the item ID:\n%s", smallJoined)
	}
	if !strings.Contains(smallJoined, "[y]") {
		t.Fatalf("small confirm modal lost the action footer:\n%s", smallJoined)
	}

	// Extremely long prompts must stay bounded; the item ID lives in the
	// title so it survives prompt ellipsization.
	extreme := modals.RenderConfirmModal(theme, "Merge PR #142", "Merge PR #142: "+strings.Repeat("this change touches the renderer carefully ", 40), 80, 24)
	extremeLines := assertModalBounds(t, extreme, 76, 22)
	extremeJoined := strings.Join(extremeLines, "\n")
	if !strings.Contains(extremeJoined, "#142") {
		t.Fatalf("extreme confirm modal lost the item ID:\n%s", extremeJoined)
	}
	if !strings.Contains(extremeJoined, "[y]") {
		t.Fatalf("extreme confirm modal lost the action footer:\n%s", extremeJoined)
	}
}

func TestModals_HelpPageIndexIsClamped(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	for _, tc := range []struct {
		name string
		page int
		want string
	}{
		{name: "negative", page: -5, want: "Page 1/3"},
		{name: "past-last", page: 99, want: "Page 3/3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := modals.RenderHelpModal(theme, 80, 24, tc.page)
			lines := assertModalBounds(t, out, 76, 22)
			joined := strings.Join(lines, "\n")
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("page %d was not clamped to %q:\n%s", tc.page, tc.want, joined)
			}
		})
	}

	wide := modals.RenderHelpModal(theme, 160, 45, -5)
	wideLines := assertModalBounds(t, wide, 156, 43)
	wideJoined := strings.Join(wideLines, "\n")
	if strings.Contains(wideJoined, "Page ") {
		t.Fatalf("out-of-range page made the wide help modal paginate:\n%s", wideJoined)
	}
	for _, group := range []string{
		"NAVIGATION",
		"VIEWS & SECTIONS",
		"PANELS & FOCUS",
		"PREVIEW PANE",
		"ACTIONS",
		"MODALS & GLOBAL",
	} {
		if !strings.Contains(wideJoined, group) {
			t.Fatalf("wide help modal missing group %q after out-of-range page:\n%s", group, wideJoined)
		}
	}
}

func TestModals_FitRemainingSupportedSizes(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	accounts := testAccounts(25)

	for _, size := range []struct{ w, h int }{
		{100, 30},
		{160, 45},
	} {
		t.Run(fmt.Sprintf("%dx%d", size.w, size.h), func(t *testing.T) {
			maxW := size.w - 4
			maxH := size.h - 2

			for page := 0; page < 3; page++ {
				out := modals.RenderHelpModal(theme, size.w, size.h, page)
				lines := assertModalBounds(t, out, maxW, maxH)
				joined := strings.Join(lines, "\n")
				if size.w < 120 {
					if !strings.Contains(joined, fmt.Sprintf("Page %d/3", page+1)) {
						t.Fatalf("help page %d missing footer:\n%s", page+1, joined)
					}
				} else {
					for _, group := range []string{
						"NAVIGATION",
						"VIEWS & SECTIONS",
						"PANELS & FOCUS",
						"PREVIEW PANE",
						"ACTIONS",
						"MODALS & GLOBAL",
					} {
						if !strings.Contains(joined, group) {
							t.Fatalf("wide help modal missing group %q:\n%s", group, joined)
						}
					}
				}
			}

			accountModal := modals.RenderAccountModal(theme, accounts, "acc-01", false, len(accounts), size.w, size.h)
			accountLines := assertModalBounds(t, accountModal, maxW, maxH)
			var selected []string
			for _, line := range accountLines {
				if strings.Contains(line, ">") {
					selected = append(selected, line)
				}
			}
			if len(selected) != 1 || !strings.Contains(selected[0], "Account 25") {
				t.Fatalf("last account was not the single selected row:\n%s", strings.Join(accountLines, "\n"))
			}

			prompt := "Merge PR #142: " + strings.Repeat("this change touches the renderer carefully ", 6) + "?"
			confirm := modals.RenderConfirmModal(theme, "Confirmation Required", prompt, size.w, size.h)
			confirmLines := assertModalBounds(t, confirm, maxW, maxH)
			confirmJoined := strings.Join(confirmLines, "\n")
			if !strings.Contains(confirmJoined, "#142") || !strings.Contains(confirmJoined, "[y]") {
				t.Fatalf("confirmation modal lost required content:\n%s", confirmJoined)
			}
		})
	}
}
