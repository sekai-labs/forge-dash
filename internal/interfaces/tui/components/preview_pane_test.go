package components_test

import (
	"testing"
	"time"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestPreviewPane_Render(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	pane := components.NewPreviewPane(theme)
	pane.SetSize(80, 25)

	cr := &change.ChangeRequest{
		ResourceID: forge.ResourceID{
			Provider: account.ForgeGitHub,
		},
		Title:        "Feature: responsive table",
		SourceBranch: "feature/responsive-table",
		TargetBranch: "main",
		Author: shared.User{
			Username: "octocat",
		},
		Stats: change.DiffStats{
			Additions: 42,
			Deletions: 12,
			Files:     3,
		},
		Labels: []shared.Label{
			{Name: "enhancement"},
		},
		Body: "Responsive table implementation description",
	}

	diffs := []shared.DiffFile{
		{
			OldPath:   "foo.go",
			NewPath:   "foo.go",
			Additions: 10,
			Deletions: 2,
			Patch:     "@@ -1,3 +1,5 @@\n+added line\n-removed line\n context line",
		},
	}

	commits := []shared.Commit{
		{
			ShortSHA: "abcdef1",
			Message:  "feat: improve column responsiveness",
			Author:   shared.User{Username: "octocat"},
			Date:     time.Now().Add(-1 * time.Hour),
		},
	}

	checks := []shared.CheckRun{
		{
			Name:       "build",
			Status:     "completed",
			Conclusion: shared.CheckConclusionSuccess,
		},
	}

	comments := []shared.Comment{
		{
			Author:    shared.User{Username: "reviewer"},
			Body:      "Looks great!\nShip it.",
			CreatedAt: time.Now().Add(-30 * time.Minute),
		},
	}

	for _, tab := range []components.PreviewTab{
		components.PreviewTabOverview,
		components.PreviewTabFiles,
		components.PreviewTabCommits,
		components.PreviewTabChecks,
		components.PreviewTabActivity,
	} {
		pane.SetTab(tab)
		out := pane.RenderChange(cr, diffs, commits, checks, comments)
		if out == "" {
			t.Fatalf("expected non-empty output for tab %v", tab)
		}
	}

	iss := &issue.Issue{
		ResourceID: forge.ResourceID{
			Provider: account.ForgeGitHub,
		},
		Title: "Issue title",
		Author: shared.User{
			Username: "user",
		},
		Body: "Issue description",
	}

	outIssue := pane.RenderIssue(iss, comments)
	if outIssue == "" {
		t.Fatalf("expected non-empty output for issue")
	}
}
