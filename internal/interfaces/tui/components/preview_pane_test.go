package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func demoChangeRequest() *change.ChangeRequest {
	return &change.ChangeRequest{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         "github.com",
			RepoOwner:    "charmbracelet",
			RepoName:     "bubbletea",
			ResourceType: forge.ResourceTypePR,
			NativeID:     "142",
		},
		Number:       142,
		Title:        "Feature: responsive table with an intentionally long title",
		SourceBranch: "feature/responsive-table",
		TargetBranch: "main",
		Author:       shared.User{Username: "octocat"},
		Stats:        change.DiffStats{Additions: 42, Deletions: 12, Files: 3},
		Labels:       []shared.Label{{Name: "enhancement"}},
		Body:         strings.Repeat("long description word ", 120),
	}
}

func demoIssue() *issue.Issue {
	return &issue.Issue{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			RepoOwner:    "charmbracelet",
			RepoName:     "bubbletea",
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     "305",
		},
		Number: 305,
		Title:  "Windows terminal cursor flicker during high-speed updates",
		Author: shared.User{Username: "win_dev"},
		Body:   strings.Repeat("long issue description word ", 120),
	}
}

func demoDiffs(long bool) []shared.DiffFile {
	patch := "@@ -1,3 +1,5 @@\n+added line\n-removed line\n context line"
	if long {
		patch = "+" + strings.Repeat("x", 300) + "\n" + strings.Repeat("long patch context ", 40)
	}
	return []shared.DiffFile{{
		OldPath:   "foo.go",
		NewPath:   "very/long/path/that/needs/truncation/foo.go",
		Additions: 10,
		Deletions: 2,
		Patch:     patch,
	}}
}

func demoCommits() []shared.Commit {
	return []shared.Commit{{
		ShortSHA: "abcdef1",
		Message:  strings.Repeat("improve column responsiveness ", 12),
		Author:   shared.User{Username: "octocat_with_a_long_username"},
		Date:     time.Now().Add(-time.Hour),
	}}
}

func demoChecks() []shared.CheckRun {
	return []shared.CheckRun{{
		Name:       strings.Repeat("build-and-integration-", 10),
		Status:     "completed",
		Conclusion: shared.CheckConclusionSuccess,
	}}
}

func demoComments() []shared.Comment {
	return []shared.Comment{{
		Author:    shared.User{Username: "reviewer_with_long_name"},
		Body:      strings.Repeat("Looks great! Ship it. ", 40),
		CreatedAt: time.Now().Add(-30 * time.Minute),
	}}
}

func previewLines(t *testing.T, rendered string) []string {
	t.Helper()
	stripped := ansi.Strip(rendered)
	stripped = strings.TrimSuffix(stripped, "\n")
	if stripped == "" {
		t.Fatalf("expected rendered preview output")
	}
	return strings.Split(stripped, "\n")
}

func assertPreviewWidths(t *testing.T, lines []string, width int) {
	t.Helper()
	for i, line := range lines {
		if got := ansi.StringWidth(line); got > width {
			t.Fatalf("preview line %d is %d cells wide, exceeds %d: %q", i, got, width, line)
		}
	}
}

func TestPreviewPane_RenderAllTabsStaysBounded(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	cr := demoChangeRequest()
	iss := demoIssue()

	for _, size := range []struct{ w, h int }{
		{40, 8},
		{60, 12},
		{80, 25},
		{120, 20},
	} {
		for _, tab := range []components.PreviewTab{
			components.PreviewTabOverview,
			components.PreviewTabActivity,
			components.PreviewTabCommits,
			components.PreviewTabChecks,
			components.PreviewTabFiles,
		} {
			pane := components.NewPreviewPane(theme)
			pane.SetSize(size.w, size.h)
			pane.SetTab(tab)

			lines := previewLines(t, pane.RenderChange(cr, demoDiffs(true), demoCommits(), demoChecks(), demoComments()))
			assertPreviewWidths(t, lines, size.w)
			if len(lines) > size.h {
				t.Fatalf("%dx%d tab %v: rendered %d lines, exceeds height", size.w, size.h, tab, len(lines))
			}
		}

		pane := components.NewPreviewPane(theme)
		pane.SetSize(size.w, size.h)
		lines := previewLines(t, pane.RenderIssue(iss, demoComments()))
		assertPreviewWidths(t, lines, size.w)
		if len(lines) > size.h {
			t.Fatalf("%dx%d issue: rendered %d lines, exceeds height", size.w, size.h, len(lines))
		}
	}
}

func TestPreviewPane_LongContentFillsAssignedHeight(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	cr := demoChangeRequest()
	iss := demoIssue()

	pane := components.NewPreviewPane(theme)
	pane.SetSize(50, 12)
	pane.SetTab(components.PreviewTabOverview)
	changeView := pane.RenderChange(cr, nil, nil, nil, nil)
	lines := previewLines(t, changeView)
	if len(lines) != 12 {
		t.Fatalf("expected tab bar plus height-2 wrapped body rows (12 lines), got %d", len(lines))
	}
	assertPreviewWidths(t, lines, 50)
	if again := pane.RenderChange(cr, nil, nil, nil, nil); again != changeView {
		t.Fatalf("repeated RenderChange changed the output")
	}

	issuePane := components.NewPreviewPane(theme)
	issuePane.SetSize(50, 12)
	issueView := issuePane.RenderIssue(iss, nil)
	issueLines := previewLines(t, issueView)
	if len(issueLines) != 12 {
		t.Fatalf("expected issue preview to fill 12 lines, got %d", len(issueLines))
	}
	assertPreviewWidths(t, issueLines, 50)
	if again := issuePane.RenderIssue(iss, nil); again != issueView {
		t.Fatalf("repeated RenderIssue changed the output")
	}
}

func TestPreviewPane_RenderingDoesNotChangeScrollOffset(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(60, 12)
	pane.SetTab(components.PreviewTabOverview)
	cr := demoChangeRequest()

	initial := pane.RenderChange(cr, nil, nil, nil, nil)
	for range 3 {
		if got := pane.RenderChange(cr, nil, nil, nil, nil); got != initial {
			t.Fatalf("rendering changed the scroll offset")
		}
	}

	pane.ScrollDown()
	scrolled := pane.RenderChange(cr, nil, nil, nil, nil)
	if scrolled == initial {
		t.Fatalf("expected scrolling to change the visible body")
	}
	if again := pane.RenderChange(cr, nil, nil, nil, nil); again != scrolled {
		t.Fatalf("rendering accumulated scroll")
	}

	pane.ScrollUp()
	if got := pane.RenderChange(cr, nil, nil, nil, nil); got != initial {
		t.Fatalf("scrolling back to the top did not restore the initial body")
	}

	pane.ScrollDown()
	pane.ScrollDown()
	if again := pane.RenderChange(cr, nil, nil, nil, nil); again == initial {
		t.Fatalf("expected repeated scrolling to move the body")
	}
	for range 5 {
		pane.ScrollUp()
	}
	if got := pane.RenderChange(cr, nil, nil, nil, nil); got != initial {
		t.Fatalf("scrolling past the top did not restore the initial body")
	}
}

func TestPreviewPane_TabCycleOrder(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(120, 20)

	if pane.ActiveTab() != components.PreviewTabOverview {
		t.Fatalf("expected initial tab Overview, got %v", pane.ActiveTab())
	}

	wantForward := []components.PreviewTab{
		components.PreviewTabActivity,
		components.PreviewTabCommits,
		components.PreviewTabChecks,
		components.PreviewTabFiles,
		components.PreviewTabOverview,
	}
	for i, want := range wantForward {
		pane.NextTab()
		if pane.ActiveTab() != want {
			t.Fatalf("NextTab step %d: expected %v, got %v", i+1, want, pane.ActiveTab())
		}
	}

	wantBackward := []components.PreviewTab{
		components.PreviewTabFiles,
		components.PreviewTabChecks,
		components.PreviewTabCommits,
		components.PreviewTabActivity,
		components.PreviewTabOverview,
	}
	for i, want := range wantBackward {
		pane.PrevTab()
		if pane.ActiveTab() != want {
			t.Fatalf("PrevTab step %d: expected %v, got %v", i+1, want, pane.ActiveTab())
		}
	}

	pane.SetSize(120, 10)
	pane.SetTab(components.PreviewTabOverview)
	bar := previewLines(t, pane.RenderChange(demoChangeRequest(), demoDiffs(false), demoCommits(), demoChecks(), demoComments()))[0]
	labels := []string{"Overview", "Activity", "Commits", "Checks", "Files & Diff"}
	last := -1
	for _, label := range labels {
		idx := strings.Index(bar, label)
		if idx < 0 {
			t.Fatalf("wide tab bar missing %q: %q", label, bar)
		}
		if idx <= last {
			t.Fatalf("wide tab bar out of order at %q: %q", label, bar)
		}
		last = idx
	}
}

func TestPreviewPane_IssueModeRestrictsTabs(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(80, 12)

	pane.SetTab(components.PreviewTabFiles)
	pane.SetIssueMode(true)
	if !pane.IssueMode() {
		t.Fatalf("expected issue mode")
	}
	if pane.ActiveTab() != components.PreviewTabOverview {
		t.Fatalf("issue mode must clamp PR-only tabs to Overview, got %v", pane.ActiveTab())
	}

	wantIssueCycle := []components.PreviewTab{
		components.PreviewTabActivity,
		components.PreviewTabOverview,
		components.PreviewTabActivity,
		components.PreviewTabOverview,
	}
	for i, want := range wantIssueCycle {
		pane.NextTab()
		if pane.ActiveTab() != want {
			t.Fatalf("issue NextTab step %d: expected %v, got %v", i+1, want, pane.ActiveTab())
		}
	}
	for i, want := range []components.PreviewTab{
		components.PreviewTabActivity,
		components.PreviewTabOverview,
	} {
		pane.PrevTab()
		if pane.ActiveTab() != want {
			t.Fatalf("issue PrevTab step %d: expected %v, got %v", i+1, want, pane.ActiveTab())
		}
	}

	pane.SetTab(components.PreviewTabCommits)
	if pane.ActiveTab() != components.PreviewTabOverview {
		t.Fatalf("SetTab must clamp PR-only tabs in issue mode, got %v", pane.ActiveTab())
	}
	pane.SetTab(components.PreviewTabActivity)
	if pane.ActiveTab() != components.PreviewTabActivity {
		t.Fatalf("SetTab must allow Activity in issue mode, got %v", pane.ActiveTab())
	}

	bar := previewLines(t, pane.RenderIssue(demoIssue(), demoComments()))[0]
	if !strings.Contains(bar, "Overview") || !strings.Contains(bar, "Activity") {
		t.Fatalf("issue tab bar missing Overview/Activity: %q", bar)
	}
	for _, prOnly := range []string{"Commits", "Checks", "Files & Diff", "Diff"} {
		if strings.Contains(bar, prOnly) {
			t.Fatalf("issue tab bar leaked PR tab %q: %q", prOnly, bar)
		}
	}

	pane.SetIssueMode(false)
	pane.SetTab(components.PreviewTabFiles)
	if pane.ActiveTab() != components.PreviewTabFiles {
		t.Fatalf("leaving issue mode must restore PR tabs, got %v", pane.ActiveTab())
	}
}

func TestPreviewPane_NarrowTabLabels(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(40, 10)
	pane.SetTab(components.PreviewTabOverview)

	bar := previewLines(t, pane.RenderChange(demoChangeRequest(), demoDiffs(false), demoCommits(), demoChecks(), demoComments()))[0]
	if strings.Contains(bar, "Overview") || strings.Contains(bar, "Files & Diff") {
		t.Fatalf("narrow tab bar should use short labels: %q", bar)
	}
	shortLabels := []string{"Ov", "Act", "Comm", "CI", "Diff"}
	last := -1
	for _, label := range shortLabels {
		idx := strings.Index(bar, label)
		if idx < 0 {
			t.Fatalf("narrow tab bar missing %q: %q", label, bar)
		}
		if idx <= last {
			t.Fatalf("narrow tab bar out of order at %q: %q", label, bar)
		}
		last = idx
	}
}

func TestPreviewPane_DiffLinesAreTruncatedHorizontally(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(40, 10)
	pane.SetTab(components.PreviewTabFiles)

	lines := previewLines(t, pane.RenderChange(demoChangeRequest(), demoDiffs(true), nil, nil, nil))
	assertPreviewWidths(t, lines, 40)
	if len(lines) > 10 {
		t.Fatalf("diff preview rendered %d lines, exceeds height", len(lines))
	}
}
