package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestItemTable_RenderChanges(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	tbl := components.NewItemTable(theme)
	tbl.SetSize(120, 20)
	tbl.SetFocused(true)

	changes := []change.ChangeRequest{
		{
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
			Author: shared.User{
				Username: "octocat",
			},
			UpdatedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	out := tbl.RenderChanges(changes)
	if out == "" {
		t.Fatalf("expected non-empty table output")
	}

	tbl.MoveDown(len(changes))
	if tbl.Cursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", tbl.Cursor())
	}
}

func TestItemTable_RenderIssuesAndNotifications(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	tbl := components.NewItemTable(theme)
	tbl.SetSize(100, 15)

	issues := []issue.Issue{
		{
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
			Title:     "Crash on terminal resize under 80x24",
			State:     issue.IssueStateOpen,
			Author: shared.User{
				Username: "gitlab-user",
			},
			Comments:  3,
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	outIssues := tbl.RenderIssues(issues)
	if outIssues == "" {
		t.Fatalf("expected non-empty issues output")
	}

	notifs := []notification.Notification{
		{
			ID: "notif-1",
			ResourceID: forge.ResourceID{
				Provider: account.ForgeGitHub,
			},
			RepoOwner:    "sekai-labs",
			RepoName:     "forge-dash",
			SubjectTitle: "Review requested on PR #101",
			Reason:       "review_requested",
			Unread:       true,
			UpdatedAt:    time.Now().Add(-5 * time.Minute),
		},
	}

	outNotifs := tbl.RenderNotifications(notifs)
	if outNotifs == "" {
		t.Fatalf("expected non-empty notifications output")
	}
}

func TestItemTable_HeaderNeverWraps(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	changes := []change.ChangeRequest{
		{
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
			Author: shared.User{
				Username: "octocat",
			},
			UpdatedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	for _, w := range []int{40, 55, 60, 70, 75, 80, 100, 120, 160} {
		tbl := components.NewItemTable(theme)
		tbl.SetSize(w, 10)
		out := tbl.RenderChanges(changes)
		boxed := theme.Box.Width(w).Render(out)
		lines := strings.Split(boxed, "\n")

		if len(lines) < 3 {
			t.Fatalf("width %d: expected at least 3 lines in boxed table, got %d", w, len(lines))
		}

		if !strings.Contains(lines[2], "─") {
			t.Errorf("width %d: line 2 expected to be separator line with ─, but got: %q (line 1 was %q)", w, lines[2], lines[1])
		}
	}
}
