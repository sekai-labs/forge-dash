package components_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/domain/account"
	"github.com/sekai-labs/forge-dash/internal/domain/change"
	"github.com/sekai-labs/forge-dash/internal/domain/forge"
	"github.com/sekai-labs/forge-dash/internal/domain/issue"
	"github.com/sekai-labs/forge-dash/internal/domain/notification"
	"github.com/sekai-labs/forge-dash/internal/domain/shared"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/components"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
)

func tableChanges(n int) []change.ChangeRequest {
	now := time.Now().Add(-10 * time.Minute)
	changes := make([]change.ChangeRequest, 0, n)
	for i := range n {
		changes = append(changes, change.ChangeRequest{
			ResourceID: forge.ResourceID{
				Provider:     account.ForgeGitHub,
				Host:         "github.com",
				RepoOwner:    "sekai-labs",
				RepoName:     "forge-dash",
				ResourceType: forge.ResourceTypePR,
				NativeID:     "#" + strconv.Itoa(100+i),
			},
			Number:       100 + i,
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			Title:        "Item " + strconv.Itoa(i) + " unique title",
			State:        change.ChangeStateOpen,
			CIStatus:     change.CIStatusPassing,
			ReviewStatus: change.ReviewStatusApproved,
			Author:       shared.User{Username: "octocat"},
			UpdatedAt:    now,
		})
	}
	return changes
}

func tableLines(t *testing.T, rendered string) []string {
	t.Helper()
	stripped := ansi.Strip(rendered)
	stripped = strings.TrimSuffix(stripped, "\n")
	if stripped == "" {
		t.Fatalf("expected rendered table output")
	}
	return strings.Split(stripped, "\n")
}

func assertTableWidths(t *testing.T, lines []string, width int) {
	t.Helper()
	for i, line := range lines {
		if got := ansi.StringWidth(line); got > width {
			t.Fatalf("table line %d is %d cells wide, exceeds %d: %q", i, got, width, line)
		}
	}
}

type columnSegment struct {
	name       string
	start, end int
}

func headerColumnSegments(header, row string, names []string) []columnSegment {
	var segments []columnSegment
	for _, name := range names {
		idx := strings.Index(header, name)
		if idx < 0 {
			continue
		}
		segments = append(segments, columnSegment{name: name, start: idx})
	}
	rowWidth := ansi.StringWidth(row)
	for i := range segments {
		if i+1 < len(segments) {
			segments[i].end = segments[i+1].start
		} else {
			segments[i].end = rowWidth
		}
	}
	return segments
}

func displaySlice(line string, start, width int) string {
	if start < 0 || width <= 0 {
		return ""
	}
	if start > 0 {
		prefix := ansi.Truncate(line, start, "")
		if ansi.StringWidth(prefix) < start {
			return ""
		}
		line = line[len(prefix):]
	}
	return ansi.Truncate(line, width, "")
}

func assertColumnSegments(t *testing.T, label string, segments []columnSegment, expected []string, row string, want map[string]string) {
	t.Helper()
	if len(segments) != len(expected) {
		names := make([]string, len(segments))
		for i, seg := range segments {
			names[i] = seg.name
		}
		t.Fatalf("%s: expected visible columns %v, got %v", label, expected, names)
	}
	wanted := make(map[string]bool, len(expected))
	for _, name := range expected {
		wanted[name] = true
	}
	for _, seg := range segments {
		if !wanted[seg.name] {
			t.Fatalf("%s: unexpected visible column %s", label, seg.name)
		}
		expectedVal, ok := want[seg.name]
		if !ok {
			continue
		}
		cell := displaySlice(row, seg.start, seg.end-seg.start)
		if !strings.Contains(cell, expectedVal) {
			t.Fatalf("%s: column %s does not contain %q, cell: %q (row/header mapping shifted)", label, seg.name, expectedVal, cell)
		}
	}
}

func TestItemTable_ClampTracksCursorAcrossResizeAndDataChange(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	tbl := components.NewItemTable(theme)
	tbl.SetSize(120, 5)

	changes := tableChanges(10)
	tbl.MoveBottom(len(changes))
	tbl.Clamp(len(changes))

	if tbl.Cursor() != 9 {
		t.Fatalf("expected cursor at last row 9, got %d", tbl.Cursor())
	}
	out := tbl.RenderChanges(changes)
	lines := tableLines(t, out)
	assertTableWidths(t, lines, 120)
	if len(lines) != 5 {
		t.Fatalf("expected header+separator+3 visible rows at height 5, got %d lines", len(lines))
	}
	marker := ""
	for _, line := range lines {
		if strings.Contains(line, "▎") {
			if marker != "" {
				t.Fatalf("multiple selected rows rendered")
			}
			marker = line
		}
	}
	if !strings.Contains(marker, "#109") {
		t.Fatalf("expected last row to stay visible after clamp, marker line: %q", marker)
	}

	// Growing the table must not move the cursor.
	tbl.SetSize(120, 8)
	tbl.Clamp(10)
	if tbl.Cursor() != 9 {
		t.Fatalf("resize moved the cursor to %d", tbl.Cursor())
	}
	lines = tableLines(t, tbl.RenderChanges(changes))
	if len(lines) > 8 {
		t.Fatalf("expected at most 8 lines at height 8, got %d", len(lines))
	}
	if marker := func() string {
		for _, line := range lines {
			if strings.Contains(line, "▎") {
				return line
			}
		}
		return ""
	}(); !strings.Contains(marker, "#109") {
		t.Fatalf("expected selection to stay visible after resize, marker line: %q", marker)
	}

	// Shrinking the data set must clamp the cursor into range.
	tbl.Clamp(3)
	if tbl.Cursor() != 2 {
		t.Fatalf("expected cursor clamped to 2, got %d", tbl.Cursor())
	}
	lines = tableLines(t, tbl.RenderChanges(changes[:3]))
	if marker := func() string {
		for _, line := range lines {
			if strings.Contains(line, "▎") {
				return line
			}
		}
		return ""
	}(); !strings.Contains(marker, "#102") {
		t.Fatalf("expected clamped cursor to render #102, marker line: %q", marker)
	}

	tbl.Clamp(0)
	if tbl.Cursor() != 0 {
		t.Fatalf("expected empty data to reset cursor, got %d", tbl.Cursor())
	}
	empty := tbl.RenderChanges(nil)
	emptyLines := tableLines(t, empty)
	assertTableWidths(t, emptyLines, 120)
	if !strings.Contains(strings.Join(emptyLines, "\n"), "No items in this section.") {
		t.Fatalf("expected the explicit empty-state copy, got:\n%s", strings.Join(emptyLines, "\n"))
	}
}

func TestItemTable_RenderingDoesNotMutateCursor(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	tbl := components.NewItemTable(theme)
	tbl.SetSize(120, 4)

	changes := tableChanges(5)
	for range 4 {
		tbl.MoveDown(len(changes))
	}
	if tbl.Cursor() != 4 {
		t.Fatalf("expected cursor 4, got %d", tbl.Cursor())
	}

	first := tbl.RenderChanges(changes)
	for range 3 {
		if got := tbl.RenderChanges(changes); got != first {
			t.Fatalf("repeated table rendering changed the output")
		}
		if tbl.Cursor() != 4 {
			t.Fatalf("rendering moved the cursor to %d", tbl.Cursor())
		}
	}

	issues := []issue.Issue{{
		ResourceID: forge.ResourceID{Provider: account.ForgeGitLab, ResourceType: forge.ResourceTypeIssue},
		Number:     7,
		RepoOwner:  "sekai-labs",
		RepoName:   "forge-dash",
		Title:      "Issue title",
		State:      issue.IssueStateOpen,
		Author:     shared.User{Username: "gitlab-user"},
		UpdatedAt:  time.Now().Add(-time.Hour),
	}}
	_ = tbl.RenderIssues(issues)
	if tbl.Cursor() != 4 {
		t.Fatalf("rendering issues moved the cursor to %d", tbl.Cursor())
	}

	notifs := []notification.Notification{{
		ID:           "notif-1",
		ResourceID:   forge.ResourceID{Provider: account.ForgeGitHub},
		RepoOwner:    "sekai-labs",
		RepoName:     "forge-dash",
		SubjectTitle: "Review requested on PR #101",
		Reason:       "review_requested",
		Unread:       true,
		UpdatedAt:    time.Now().Add(-5 * time.Minute),
	}}
	_ = tbl.RenderNotifications(notifs)
	if tbl.Cursor() != 4 {
		t.Fatalf("rendering notifications moved the cursor to %d", tbl.Cursor())
	}
}

func TestItemTable_ColumnValuesStayAlignedAfterPruning(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	changes := []change.ChangeRequest{{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         "github.com",
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			ResourceType: forge.ResourceTypePR,
			NativeID:     "101",
		},
		Number:       101,
		RepoOwner:    "sekai-labs",
		RepoName:     "forge-dash",
		Title:        "Support opentui modern theme & unicode string handling",
		State:        change.ChangeStateOpen,
		CIStatus:     change.CIStatusPassing,
		ReviewStatus: change.ReviewStatusApproved,
		Author:       shared.User{Username: "octocat"},
		UpdatedAt:    time.Now().Add(-2 * time.Hour),
	}}

	headers := []string{"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "CI", "REV", "TIME"}
	want := map[string]string{
		"PROV":   "[GH]",
		"REPO":   "sekai-labs",
		"ID":     "#101",
		"TITLE":  "Sup",
		"AUTHOR": "@octocat",
		"STATE":  "Open",
		"CI":     "✓",
		"REV":    "✓",
		"TIME":   "2h",
	}

	// Widths that prune secondary columns. Interior pruning must not shift
	// the values that follow the removed column under the wrong header.
	expectedColumns := map[int][]string{
		40:  {"PROV", "REPO", "ID", "TITLE"},
		55:  {"PROV", "REPO", "ID", "TITLE", "STATE", "TIME"},
		58:  {"PROV", "REPO", "ID", "TITLE", "STATE", "TIME"},
		70:  {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "TIME"},
		75:  {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "TIME"},
		90:  {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "CI", "REV", "TIME"},
		120: {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "CI", "REV", "TIME"},
		160: {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "CI", "REV", "TIME"},
	}

	for _, width := range []int{40, 55, 58, 70, 75, 90, 120, 160} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(width, 6)
		lines := tableLines(t, tbl.RenderChanges(changes))
		assertTableWidths(t, lines, width)
		if len(lines) < 3 {
			t.Fatalf("width %d: expected header, separator, and a row, got %d lines", width, len(lines))
		}

		header := lines[0]
		row := strings.ReplaceAll(lines[2], "▎", " ")
		segments := headerColumnSegments(header, row, headers)
		assertColumnSegments(t, fmt.Sprintf("width %d", width), segments, expectedColumns[width], row, want)
	}
}

func TestItemTable_HeaderNeverWraps(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	changes := []change.ChangeRequest{{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         "github.com",
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			ResourceType: forge.ResourceTypePR,
			NativeID:     "12",
		},
		Number:       12,
		RepoOwner:    "sekai-labs",
		RepoName:     "forge-dash",
		Title:        "feat: unified status bar with responsive design and fast render",
		State:        change.ChangeStateOpen,
		CIStatus:     change.CIStatusPassing,
		ReviewStatus: change.ReviewStatusApproved,
		Author:       shared.User{Username: "octocat"},
		UpdatedAt:    time.Now().Add(-10 * time.Minute),
	}}

	for _, width := range []int{40, 55, 60, 70, 75, 80, 100, 120, 160} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(width, 10)
		lines := tableLines(t, tbl.RenderChanges(changes))
		assertTableWidths(t, lines, width)

		if len(lines) < 3 {
			t.Fatalf("width %d: expected at least 3 lines, got %d", width, len(lines))
		}
		if !strings.Contains(lines[1], "─") {
			t.Errorf("width %d: expected separator line, got %q (header %q)", width, lines[1], lines[0])
		}
	}
}

func TestItemTable_IssuesAndNotificationsStayBounded(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	longTitle := strings.Repeat("very long issue title ", 8)
	issues := []issue.Issue{{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         "gitlab.com",
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     "42",
		},
		Number:    42,
		RepoOwner: "sekai-labs",
		RepoName:  "forge-dash",
		Title:     longTitle,
		State:     issue.IssueStateOpen,
		Author:    shared.User{Username: "gitlab-user"},
		Comments:  3,
		UpdatedAt: time.Now().Add(-time.Hour),
	}}
	notifs := []notification.Notification{{
		ID:           "notif-1",
		ResourceID:   forge.ResourceID{Provider: account.ForgeGitHub},
		RepoOwner:    "sekai-labs",
		RepoName:     "forge-dash",
		SubjectTitle: longTitle,
		Reason:       "review_requested",
		Unread:       true,
		UpdatedAt:    time.Now().Add(-5 * time.Minute),
	}}

	for _, width := range []int{40, 60, 80, 120} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(width, 8)

		issueLines := tableLines(t, tbl.RenderIssues(issues))
		assertTableWidths(t, issueLines, width)
		if len(issueLines) > 8 {
			t.Fatalf("width %d: issues table exceeds assigned height: %d", width, len(issueLines))
		}

		notifLines := tableLines(t, tbl.RenderNotifications(notifs))
		assertTableWidths(t, notifLines, width)
		if len(notifLines) > 8 {
			t.Fatalf("width %d: notifications table exceeds assigned height: %d", width, len(notifLines))
		}
	}

	tbl := components.NewItemTable(theme)
	tbl.SetSize(40, 5)
	empty := tbl.RenderChanges(nil)
	emptyLines := tableLines(t, empty)
	assertTableWidths(t, emptyLines, 40)
	if !strings.Contains(strings.Join(emptyLines, "\n"), "No items in this section.") {
		t.Fatalf("expected the explicit empty-state copy, got:\n%s", strings.Join(emptyLines, "\n"))
	}
}

func TestItemTable_IssueCellMappingAfterInteriorPrune(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	issues := []issue.Issue{{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         "gitlab.com",
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     "42",
		},
		Number:    42,
		RepoOwner: "sekai-labs",
		RepoName:  "forge-dash",
		Title:     "Windows terminal cursor flicker during high-speed updates",
		State:     issue.IssueStateOpen,
		Author:    shared.User{Username: "gitlab-user"},
		Comments:  3,
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	}}

	headers := []string{"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "COMMENTS", "TIME"}
	want := map[string]string{
		"PROV":     "[GL]",
		"REPO":     "sekai-labs",
		"ID":       "#42",
		"TITLE":    "Windo",
		"AUTHOR":   "gitla",
		"STATE":    "Open",
		"COMMENTS": "3",
		"TIME":     "2h",
	}

	expectedColumns := map[int][]string{
		40: {"PROV", "REPO", "ID", "TITLE"},
		55: {"PROV", "REPO", "ID", "TITLE", "STATE", "TIME"},
		70: {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "TIME"},
		75: {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "TIME"},
		90: {"PROV", "REPO", "ID", "TITLE", "AUTHOR", "STATE", "COMMENTS", "TIME"},
	}

	for _, width := range []int{40, 55, 70, 75, 90} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(width, 6)
		lines := tableLines(t, tbl.RenderIssues(issues))
		assertTableWidths(t, lines, width)
		if len(lines) < 3 {
			t.Fatalf("issues width %d: expected header, separator, and a row, got %d lines", width, len(lines))
		}

		header := lines[0]
		row := strings.ReplaceAll(lines[2], "▎", " ")
		segments := headerColumnSegments(header, row, headers)
		assertColumnSegments(t, fmt.Sprintf("issues width %d", width), segments, expectedColumns[width], row, want)
	}
}

func TestItemTable_NotificationCellMappingAfterInteriorPrune(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	notifs := []notification.Notification{{
		ID:           "notif-1",
		ResourceID:   forge.ResourceID{Provider: account.ForgeGitHub},
		RepoOwner:    "sekai-labs",
		RepoName:     "forge-dash",
		SubjectTitle: "Review requested on PR #101",
		Reason:       "review_requested",
		Unread:       true,
		UpdatedAt:    time.Now().Add(-5 * time.Minute),
	}}

	headers := []string{"PROV", "REASON", "REPO", "TITLE", "STATE", "TIME"}
	want := map[string]string{
		"PROV":   "[GH]",
		"REASON": "review",
		"REPO":   "sekai-labs",
		"TITLE":  "Review",
		"STATE":  "un",
		"TIME":   "5m",
	}

	expectedColumns := map[int][]string{
		40: {"PROV", "REPO", "TITLE", "STATE"},
		60: {"PROV", "REASON", "REPO", "TITLE", "STATE", "TIME"},
		80: {"PROV", "REASON", "REPO", "TITLE", "STATE", "TIME"},
	}

	for _, width := range []int{40, 60, 80} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(width, 6)
		lines := tableLines(t, tbl.RenderNotifications(notifs))
		assertTableWidths(t, lines, width)
		if len(lines) < 3 {
			t.Fatalf("notifications width %d: expected header, separator, and a row, got %d lines", width, len(lines))
		}

		header := lines[0]
		row := strings.ReplaceAll(lines[2], "▎", " ")
		segments := headerColumnSegments(header, row, headers)
		assertColumnSegments(t, fmt.Sprintf("notifications width %d", width), segments, expectedColumns[width], row, want)
	}
}
