package components_test

import (
	"testing"
	"time"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestItemTable_RenderChanges(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())
	tbl := components.NewItemTable(theme)
	tbl.SetSize(120, 20)

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
